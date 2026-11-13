package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/brii26/eth-iota-attestation/oracle/internal/chain"
	"github.com/brii26/eth-iota-attestation/oracle/internal/model"
	"github.com/brii26/eth-iota-attestation/oracle/internal/store"
	"github.com/brii26/eth-iota-attestation/oracle/internal/verify"
)

func setupTestServer(t *testing.T, artifactContent []byte) (*Server, *store.SQLiteIndex, *chain.MockAdapter, *httptest.Server) {
	tmpDir := t.TempDir()
	idx, err := store.NewSQLiteIndex(filepath.Join(tmpDir, "test.sqlite"))
	if err != nil {
		t.Fatalf("failed to init test sqlite: %v", err)
	}

	mockChain := chain.NewMockAdapter()

	// Mock Verdaccio Server
	verdaccioSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/demo-lib":
			doc := map[string]interface{}{
				"versions": map[string]interface{}{
					"1.4.2": map[string]interface{}{
						"dist": map[string]interface{}{
							"tarball": "http://" + r.Host + "/demo-lib/-/demo-lib-1.4.2.tgz",
						},
					},
				},
			}
			json.NewEncoder(w).Encode(doc)
		case "/demo-lib/-/demo-lib-1.4.2.tgz":
			w.Write(artifactContent)
		default:
			http.NotFound(w, r)
		}
	}))

	vClient := verify.NewVerdaccioClient(verdaccioSrv.URL)

	srv := NewServer(Config{
		Index:           idx,
		ChainAdapter:    mockChain,
		VerdaccioClient: vClient,
		MaxGrypeDBAge:   24 * time.Hour,
	})

	return srv, idx, mockChain, verdaccioSrv
}

func shaHex(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func TestAttestationAPI(t *testing.T) {
	mockArtifact := []byte("package-tarball-binary-contents")
	nowStr := time.Now().UTC().Format(time.RFC3339)

	mockSBOM := []byte(`{"bomFormat":"CycloneDX","specVersion":"1.4","components":[]}`)
	mockScan := []byte(fmt.Sprintf(`{"descriptor":{"name":"grype","version":"0.80.0","db":{"built":%q,"status":"active"}}}`, nowStr))

	srv, idx, mockChain, verdaccioSrv := setupTestServer(t, mockArtifact)
	defer idx.Close()
	defer verdaccioSrv.Close()

	validScanID := "0f8c2d3e-6b1a-4c55-9e1f-2a7b8c9d0e11"

	reqBody := model.AttestationRequest{
		Network: "ethereum-private",
		App:     "demo-lib",
		Version: "1.4.2",
		ScanID:  validScanID,
		ClaimedHashes: model.ClaimedHashes{
			Artifact: shaHex(mockArtifact),
			SBOM:     shaHex(mockSBOM),
			Scan:     shaHex(mockScan),
		},
		SBOM: mockSBOM,
		Scan: mockScan,
	}

	// 1. Success Flow
	t.Run("Valid Submission Returns 201 Created", func(t *testing.T) {
		payload, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/attestations", bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer mock-ci-token")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d. body: %s", rec.Code, rec.Body.String())
		}

		var res model.AttestationStatus
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if res.Status != "included" || res.TxRef == "" || res.RecordID == "" {
			t.Errorf("unexpected status response: %+v", res)
		}

		// Verify record is indexed in SQLite
		st, err := idx.GetStatus(req.Context(), reqBody.Network, reqBody.ScanID)
		if err != nil || st == nil || st.Status != "included" {
			t.Errorf("expected record in sqlite with status included, got %+v, err=%v", st, err)
		}

		// Verify chain adapter received record
		if len(mockChain.Records()) != 1 {
			t.Errorf("expected 1 record in chain adapter, got %d", len(mockChain.Records()))
		}
	})

	// 2. Duplicate Scan ID Rejected
	t.Run("Duplicate Scan ID Returns 409 Conflict", func(t *testing.T) {
		payload, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/attestations", bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer mock-ci-token")
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("expected status 409 for duplicate, got %d", rec.Code)
		}

		var errRes model.ErrorResponse
		json.Unmarshal(rec.Body.Bytes(), &errRes)
		if errRes.Error.Code != model.ErrCodeDuplicateScanID {
			t.Errorf("expected error code %s, got %s", model.ErrCodeDuplicateScanID, errRes.Error.Code)
		}
	})

	// 3. Claimed Hash Mismatch Rejected
	t.Run("Hash Mismatch Returns 400 Bad Request", func(t *testing.T) {
		badHashReq := reqBody
		badHashReq.ScanID = "11111111-2222-3333-4444-555555555555"
		badHashReq.ClaimedHashes.Artifact = "0000000000000000000000000000000000000000000000000000000000000000"

		payload, _ := json.Marshal(badHashReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/attestations", bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer mock-ci-token")
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 for hash mismatch, got %d", rec.Code)
		}
	})

	// 4. Stale Scan Rejected
	t.Run("Stale Scan DB Returns 422 Unprocessable Entity", func(t *testing.T) {
		staleTimeStr := time.Now().UTC().Add(-72 * time.Hour).Format(time.RFC3339)
		staleScan := []byte(fmt.Sprintf(`{"descriptor":{"name":"grype","version":"0.80.0","db":{"built":%q,"status":"active"}}}`, staleTimeStr))

		staleReq := reqBody
		staleReq.ScanID = "22222222-3333-4444-5555-666666666666"
		staleReq.Scan = staleScan
		staleReq.ClaimedHashes.Scan = shaHex(staleScan)

		payload, _ := json.Marshal(staleReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/attestations", bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer mock-ci-token")
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422 for stale scan, got %d", rec.Code)
		}

		var errRes model.ErrorResponse
		json.Unmarshal(rec.Body.Bytes(), &errRes)
		if errRes.Error.Code != model.ErrCodeStaleScan {
			t.Errorf("expected error code %s, got %s", model.ErrCodeStaleScan, errRes.Error.Code)
		}
	})

	// 5. Dependency Unavailable (Package Not Found in Verdaccio)
	t.Run("Missing Artifact Returns 502 Bad Gateway", func(t *testing.T) {
		missingReq := reqBody
		missingReq.ScanID = "33333333-4444-5555-6666-777777777777"
		missingReq.App = "non-existent-package"

		payload, _ := json.Marshal(missingReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/attestations", bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer mock-ci-token")
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadGateway {
			t.Fatalf("expected status 502 for missing artifact, got %d", rec.Code)
		}

		var errRes model.ErrorResponse
		json.Unmarshal(rec.Body.Bytes(), &errRes)
		if errRes.Error.Code != model.ErrCodeDependencyUnavailable {
			t.Errorf("expected error code %s, got %s", model.ErrCodeDependencyUnavailable, errRes.Error.Code)
		}
	})

	// 6. Unauthorized Pipeline
	t.Run("Unauthorized Pipeline Returns 403 Forbidden", func(t *testing.T) {
		unauthReq := reqBody
		unauthReq.ScanID = "44444444-5555-6666-7777-888888888888"

		payload, _ := json.Marshal(unauthReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/attestations", bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer unauthorized")
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403 for unauthorized pipeline, got %d", rec.Code)
		}
	})

	// 7. GET Attestation Status
	t.Run("GET Attestation Status Returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/attestations/ethereum-private/"+validScanID, nil)
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var res model.AttestationStatus
		json.Unmarshal(rec.Body.Bytes(), &res)
		if res.ScanID != validScanID || res.Status != "included" {
			t.Errorf("unexpected status in GET response: %+v", res)
		}
	})

	// 8. GET Non-Existent Attestation Returns 404
	t.Run("GET Non-Existent Attestation Returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/attestations/ethereum-private/non-existent-scan", nil)
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", rec.Code)
		}
	})
}
