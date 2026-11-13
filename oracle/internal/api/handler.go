package api

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/brii26/eth-iota-attestation/oracle/internal/chain"
	"github.com/brii26/eth-iota-attestation/oracle/internal/model"
	"github.com/brii26/eth-iota-attestation/oracle/internal/store"
	"github.com/brii26/eth-iota-attestation/oracle/internal/verify"
)

// Config holds runtime dependencies and options for the API server.
type Config struct {
	S3Store         *store.S3Store
	Index           *store.SQLiteIndex
	ChainAdapter    chain.Adapter
	Router          *chain.Router
	VerdaccioClient *verify.VerdaccioClient
	MaxGrypeDBAge   time.Duration
}

// Server handles HTTP requests for the oracle.
type Server struct {
	cfg    Config
	mux    *http.ServeMux
	router http.Handler
}

// NewServer sets up routes and returns an initialized Server.
func NewServer(cfg Config) *Server {
	if cfg.MaxGrypeDBAge == 0 {
		cfg.MaxGrypeDBAge = 48 * time.Hour
	}

	s := &Server{
		cfg: cfg,
		mux: http.NewServeMux(),
	}
	s.routes()
	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("POST /api/v1/attestations", s.handleCreateAttestation)
	s.mux.HandleFunc("GET /api/v1/attestations/{network}/{scanId}", s.handleGetAttestation)
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(model.ErrorResponse{
		Error: model.ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (s *Server) handleCreateAttestation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Authenticate pipeline via OIDC stub
	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	token = strings.TrimSpace(token)

	// Decode request body
	var req model.AttestationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, model.ErrCodeInvalidInput, "invalid JSON payload")
		return
	}

	// Validate inputs
	if req.Network == "" || req.App == "" || req.Version == "" || req.ScanID == "" {
		writeJSONError(w, http.StatusBadRequest, model.ErrCodeInvalidInput, "network, app, version, and scanId are required")
		return
	}

	var chainAdapter chain.Adapter
	if s.cfg.Router != nil {
		ad, err := s.cfg.Router.Get(req.Network)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, model.ErrCodeInvalidInput, fmt.Sprintf("unsupported network %q", req.Network))
			return
		}
		chainAdapter = ad
	} else if s.cfg.ChainAdapter != nil {
		chainAdapter = s.cfg.ChainAdapter
	} else {
		writeJSONError(w, http.StatusInternalServerError, model.ErrCodeInternal, "no chain adapter configured")
		return
	}
	if len(req.SBOM) == 0 || len(req.Scan) == 0 {
		writeJSONError(w, http.StatusBadRequest, model.ErrCodeInvalidInput, "sbom and scan payloads are required")
		return
	}
	if req.ClaimedHashes.Artifact == "" || req.ClaimedHashes.SBOM == "" || req.ClaimedHashes.Scan == "" {
		writeJSONError(w, http.StatusBadRequest, model.ErrCodeInvalidInput, "claimed hashes for artifact, sbom, and scan are required")
		return
	}

	// Check pipeline authorization for app
	if !verify.VerifyPipeline(token, req.App) {
		writeJSONError(w, http.StatusForbidden, model.ErrCodeUnauthorizedPipeline, "unauthorized pipeline submission")
		return
	}

	// 2. Fetch artifact from Verdaccio
	artBytes, err := s.cfg.VerdaccioClient.FetchArtifact(ctx, req.App, req.Version)
	if err != nil {
		if errors.Is(err, verify.ErrArtifactNotFound) {
			writeJSONError(w, http.StatusBadGateway, model.ErrCodeDependencyUnavailable, fmt.Sprintf("package artifact %s@%s not found in registry", req.App, req.Version))
			return
		}
		writeJSONError(w, http.StatusBadGateway, model.ErrCodeDependencyUnavailable, fmt.Sprintf("failed to fetch artifact from registry: %v", err))
		return
	}

	// 3. Compute raw SHA-256 hashes and compare with claimed hashes
	artHashBytes, calcArtHash := verify.ComputeSHA256(artBytes)
	if !verify.CompareHashes(req.ClaimedHashes.Artifact, calcArtHash) {
		writeJSONError(w, http.StatusBadRequest, model.ErrCodeInvalidInput, fmt.Sprintf("artifact hash mismatch: claimed %s, computed %s", req.ClaimedHashes.Artifact, calcArtHash))
		return
	}

	sbomHashBytes, calcSbomHash := verify.ComputeSHA256(req.SBOM)
	if !verify.CompareHashes(req.ClaimedHashes.SBOM, calcSbomHash) {
		writeJSONError(w, http.StatusBadRequest, model.ErrCodeInvalidInput, fmt.Sprintf("SBOM hash mismatch: claimed %s, computed %s", req.ClaimedHashes.SBOM, calcSbomHash))
		return
	}

	scanHashBytes, calcScanHash := verify.ComputeSHA256(req.Scan)
	if !verify.CompareHashes(req.ClaimedHashes.Scan, calcScanHash) {
		writeJSONError(w, http.StatusBadRequest, model.ErrCodeInvalidInput, fmt.Sprintf("scan hash mismatch: claimed %s, computed %s", req.ClaimedHashes.Scan, calcScanHash))
		return
	}

	// 4. Verify scan ID uniqueness in SQLite
	exists, err := s.cfg.Index.Exists(ctx, req.Network, req.ScanID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, model.ErrCodeInternal, "failed to query scan ID index")
		return
	}
	if exists {
		writeJSONError(w, http.StatusConflict, model.ErrCodeDuplicateScanID, fmt.Sprintf("scan ID %s already used on network %s", req.ScanID, req.Network))
		return
	}

	// Verify Grype DB freshness
	if _, err := verify.ValidateGrypeFreshness(req.Scan, s.cfg.MaxGrypeDBAge, time.Now().UTC()); err != nil {
		writeJSONError(w, http.StatusUnprocessableEntity, model.ErrCodeStaleScan, fmt.Sprintf("scan report invalid or stale: %v", err))
		return
	}

	// 5. Store SBOM and Scan in RustFS S3
	storageKey := fmt.Sprintf("%s/%s/%s/%s", req.Network, req.App, req.Version, req.ScanID)
	if s.cfg.S3Store != nil {
		if err := s.cfg.S3Store.PutSBOM(ctx, storageKey+".json", req.SBOM); err != nil {
			writeJSONError(w, http.StatusBadGateway, model.ErrCodeDependencyUnavailable, fmt.Sprintf("failed to store SBOM: %v", err))
			return
		}
		if err := s.cfg.S3Store.PutScan(ctx, storageKey+".json", req.Scan); err != nil {
			writeJSONError(w, http.StatusBadGateway, model.ErrCodeDependencyUnavailable, fmt.Sprintf("failed to store scan: %v", err))
			return
		}
	}

	// 6. Record attestation via Chain Adapter & index in SQLite
	now := time.Now().UTC()
	dbRecord := store.AttestationRecord{
		Network:      req.Network,
		App:          req.App,
		Version:      req.Version,
		ScanID:       req.ScanID,
		ArtifactHash: hex.EncodeToString(artHashBytes[:]),
		SBOMHash:     hex.EncodeToString(sbomHashBytes[:]),
		ScanHash:     hex.EncodeToString(scanHashBytes[:]),
		Status:       "pending",
		SubmittedAt:  now,
	}

	if err := s.cfg.Index.Save(ctx, dbRecord); err != nil {
		writeJSONError(w, http.StatusInternalServerError, model.ErrCodeInternal, fmt.Sprintf("failed to record pending attestation: %v", err))
		return
	}

	chainRec := chain.Record{
		Network:      req.Network,
		App:          req.App,
		Version:      req.Version,
		ScanID:       req.ScanID,
		ArtifactHash: artHashBytes,
		SBOMHash:     sbomHashBytes,
		ScanHash:     scanHashBytes,
	}

	receipt, err := chainAdapter.WriteRecord(ctx, chainRec)
	if err != nil {
		_ = s.cfg.Index.UpdateStatus(ctx, req.Network, req.ScanID, "failed", "", "", nil)
		writeJSONError(w, http.StatusBadGateway, model.ErrCodeDependencyUnavailable, fmt.Sprintf("failed to write record to chain: %v", err))
		return
	}

	includedAt := receipt.IncludedAt
	_ = s.cfg.Index.UpdateStatus(ctx, req.Network, req.ScanID, "included", receipt.TxRef, receipt.RecordID, &includedAt)

	writeJSON(w, http.StatusCreated, model.AttestationStatus{
		Network:     req.Network,
		ScanID:      req.ScanID,
		Status:      "included",
		TxRef:       receipt.TxRef,
		RecordID:    receipt.RecordID,
		SubmittedAt: now,
		IncludedAt:  &includedAt,
	})
}

func (s *Server) handleGetAttestation(w http.ResponseWriter, r *http.Request) {
	network := r.PathValue("network")
	scanID := r.PathValue("scanId")

	if network == "" || scanID == "" {
		writeJSONError(w, http.StatusBadRequest, model.ErrCodeInvalidInput, "network and scanId path parameters required")
		return
	}

	status, err := s.cfg.Index.GetStatus(r.Context(), network, scanID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, model.ErrCodeInternal, "failed to query attestation status")
		return
	}
	if status == nil {
		writeJSONError(w, http.StatusNotFound, model.ErrCodeNotFound, fmt.Sprintf("attestation not found for network %s, scan ID %s", network, scanID))
		return
	}

	writeJSON(w, http.StatusOK, status)
}
