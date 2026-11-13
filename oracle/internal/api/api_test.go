package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
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

	router := chain.NewRouter()
	router.Register(mockChain)

	srv := NewServer(Config{
		Index:           idx,
		Router:          router,
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

	// 9. Unsupported Network Returns 400
	t.Run("Unsupported Network Returns 400 Bad Request", func(t *testing.T) {
		badNetReq := reqBody
		badNetReq.ScanID = "55555555-6666-7777-8888-999999999999"
		badNetReq.Network = "unsupported-network"

		payload, _ := json.Marshal(badNetReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/attestations", bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer mock-ci-token")
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 for unsupported network, got %d", rec.Code)
		}
	})
}

type testBlobStore struct {
	sboms map[string][]byte
	scans map[string][]byte
}

func (m *testBlobStore) GetSBOM(ctx context.Context, key string) ([]byte, error) {
	if data, ok := m.sboms[key]; ok {
		return data, nil
	}
	return nil, errors.New("sbom not found in test blob store")
}

func (m *testBlobStore) GetScan(ctx context.Context, key string) ([]byte, error) {
	if data, ok := m.scans[key]; ok {
		return data, nil
	}
	return nil, errors.New("scan not found in test blob store")
}

func TestVerifyAPI(t *testing.T) {
	tmpDir := t.TempDir()
	idx, err := store.NewSQLiteIndex(filepath.Join(tmpDir, "test-verify.sqlite"))
	if err != nil {
		t.Fatalf("failed to init sqlite: %v", err)
	}
	defer idx.Close()

	mockChain := chain.NewMockAdapter("ethereum-private")
	router := chain.NewRouter()
	router.Register(mockChain)

	var currentArtifact []byte = []byte("valid-tarball-v1.4.2")

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
			w.Write(currentArtifact)
		default:
			http.NotFound(w, r)
		}
	}))
	defer verdaccioSrv.Close()

	blobStore := &testBlobStore{
		sboms: make(map[string][]byte),
		scans: make(map[string][]byte),
	}

	srv := NewServer(Config{
		Index:           idx,
		Router:          router,
		BlobStore:       blobStore,
		VerdaccioClient: verify.NewVerdaccioClient(verdaccioSrv.URL),
		MaxGrypeDBAge:   24 * time.Hour,
	})

	app := "demo-lib"
	version := "1.4.2"
	scanID := "scan-verify-1"
	storageKey := fmt.Sprintf("ethereum-private/%s/%s/%s.json", app, version, scanID)

	sbomBytes := []byte(`{"bomFormat":"CycloneDX"}`)
	freshTime := time.Now().UTC().Format(time.RFC3339)
	scanBytes := []byte(fmt.Sprintf(`{"descriptor":{"db":{"built":%q}}}`, freshTime))

	artHash, _ := verify.ComputeSHA256(currentArtifact)
	sbomHash, _ := verify.ComputeSHA256(sbomBytes)
	scanHash, _ := verify.ComputeSHA256(scanBytes)

	blobStore.sboms[storageKey] = sbomBytes
	blobStore.scans[storageKey] = scanBytes

	signer := "0x8a3fB9000000000000000000000000000000c19e"
	mockChain.SetRecord(chain.ChainRecord{
		RecordID:     "42",
		App:          app,
		Version:      version,
		ScanID:       scanID,
		ArtifactHash: artHash,
		SBOMHash:     sbomHash,
		ScanHash:     scanHash,
		Signer:       signer,
		Timestamp:    time.Now().UTC(),
	})

	// 1. Verify returns VALID
	t.Run("Valid verify returns 200 VALID", func(t *testing.T) {
		reqBody := model.VerifyRequest{
			Network: "ethereum-private",
			App:     app,
			Version: version,
		}
		payload, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/verify", bytes.NewReader(payload))
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. body: %s", rec.Code, rec.Body.String())
		}

		var res model.VerifyResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if res.Status != model.StatusValid || res.Stale {
			t.Errorf("expected VALID status with stale=false, got status=%s, stale=%v", res.Status, res.Stale)
		}
		if res.RecordID != "42" {
			t.Errorf("expected recordId 42, got %s", res.RecordID)
		}
	})

	// 2. Tampered artifact returns ARTIFACT_MISMATCH
	t.Run("Tampered artifact returns ARTIFACT_MISMATCH", func(t *testing.T) {
		currentArtifact = []byte("tampered-tarball-contents")
		defer func() {
			currentArtifact = []byte("valid-tarball-v1.4.2")
		}()

		reqBody := model.VerifyRequest{
			Network: "ethereum-private",
			App:     app,
			Version: version,
		}
		payload, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/verify", bytes.NewReader(payload))
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var res model.VerifyResponse
		json.Unmarshal(rec.Body.Bytes(), &res)
		if res.Status != model.StatusArtifactMismatch {
			t.Errorf("expected status %s, got %s", model.StatusArtifactMismatch, res.Status)
		}
		if res.Checks.Artifact != model.CheckMismatch {
			t.Errorf("expected artifact check to be MISMATCH, got %s", res.Checks.Artifact)
		}
	})

	// 3. Untrusted signer returns UNTRUSTED_SIGNER
	t.Run("Untrusted signer returns UNTRUSTED_SIGNER", func(t *testing.T) {
		mockChain.SetSignerAuthorized(signer, false)
		defer mockChain.SetSignerAuthorized(signer, true)

		reqBody := model.VerifyRequest{
			Network: "ethereum-private",
			App:     app,
			Version: version,
		}
		payload, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/verify", bytes.NewReader(payload))
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var res model.VerifyResponse
		json.Unmarshal(rec.Body.Bytes(), &res)
		if res.Status != model.StatusUntrustedSigner {
			t.Errorf("expected status %s, got %s", model.StatusUntrustedSigner, res.Status)
		}
		if res.Checks.SignerTrusted {
			t.Errorf("expected signerTrusted to be false")
		}
	})

	// 4. Missing record on chain returns RECORD_NOT_FOUND
	t.Run("Non-existent record returns RECORD_NOT_FOUND", func(t *testing.T) {
		reqBody := model.VerifyRequest{
			Network: "ethereum-private",
			App:     "non-existent-app",
			Version: "1.0.0",
		}
		payload, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/verify", bytes.NewReader(payload))
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var res model.VerifyResponse
		json.Unmarshal(rec.Body.Bytes(), &res)
		if res.Status != model.StatusRecordNotFound {
			t.Errorf("expected status %s, got %s", model.StatusRecordNotFound, res.Status)
		}
	})

	// 5. Missing required params returns 400 Bad Request
	t.Run("Missing app returns 400 Bad Request", func(t *testing.T) {
		reqBody := model.VerifyRequest{
			Network: "ethereum-private",
		}
		payload, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/verify", bytes.NewReader(payload))
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rec.Code)
		}
	})
}

func TestListRecordsAPI(t *testing.T) {
	tmpDir := t.TempDir()
	idx, err := store.NewSQLiteIndex(filepath.Join(tmpDir, "test-records.sqlite"))
	if err != nil {
		t.Fatalf("failed to init sqlite: %v", err)
	}
	defer idx.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	// Seed 3 records
	records := []store.AttestationRecord{
		{
			Network:      "ethereum-private",
			App:          "demo-lib",
			Version:      "1.0.0",
			ScanID:       "scan-1",
			ArtifactHash: "hash1",
			SBOMHash:     "sbom1",
			ScanHash:     "scan1",
			Status:       "included",
			SubmittedAt:  now.Add(-2 * time.Hour),
		},
		{
			Network:      "ethereum-private",
			App:          "demo-lib",
			Version:      "1.1.0",
			ScanID:       "scan-2",
			ArtifactHash: "hash2",
			SBOMHash:     "sbom2",
			ScanHash:     "scan2",
			Status:       "included",
			SubmittedAt:  now.Add(-1 * time.Hour),
		},
		{
			Network:      "ethereum-private",
			App:          "demo-lib",
			Version:      "1.2.0",
			ScanID:       "scan-3",
			ArtifactHash: "hash3",
			SBOMHash:     "sbom3",
			ScanHash:     "scan3",
			Status:       "included",
			SubmittedAt:  now,
		},
	}

	for _, r := range records {
		if err := idx.Save(ctx, r); err != nil {
			t.Fatalf("failed to save record: %v", err)
		}
	}

	srv := NewServer(Config{
		Index: idx,
	})

	// 1. List all records for demo-lib
	t.Run("List all records for demo-lib returns total 3", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/records?network=ethereum-private&app=demo-lib", nil)
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var res model.RecordHistoryResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if res.Total != 3 {
			t.Errorf("expected total 3, got %d", res.Total)
		}
		if len(res.Records) != 3 {
			t.Errorf("expected 3 records, got %d", len(res.Records))
		}
		// First record should be newest (version 1.2.0)
		if res.Records[0].Version != "1.2.0" {
			t.Errorf("expected first record to be version 1.2.0, got %s", res.Records[0].Version)
		}
	})

	// 2. Filter by version
	t.Run("Filter by version returns 1 match", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/records?network=ethereum-private&app=demo-lib&version=1.1.0", nil)
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var res model.RecordHistoryResponse
		json.Unmarshal(rec.Body.Bytes(), &res)
		if res.Total != 1 || len(res.Records) != 1 {
			t.Fatalf("expected 1 record, got total=%d len=%d", res.Total, len(res.Records))
		}
		if res.Records[0].Version != "1.1.0" {
			t.Errorf("expected version 1.1.0, got %s", res.Records[0].Version)
		}
	})

	// 3. Pagination with limit and offset
	t.Run("Pagination with limit=2 and offset=1 returns 2 records", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/records?network=ethereum-private&app=demo-lib&limit=2&offset=1", nil)
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var res model.RecordHistoryResponse
		json.Unmarshal(rec.Body.Bytes(), &res)
		if res.Total != 3 {
			t.Errorf("expected total 3, got %d", res.Total)
		}
		if len(res.Records) != 2 {
			t.Errorf("expected 2 records, got %d", len(res.Records))
		}
		if res.Records[0].Version != "1.1.0" {
			t.Errorf("expected offset 1 record to be 1.1.0, got %s", res.Records[0].Version)
		}
	})

	// 4. Missing required network or app parameter
	t.Run("Missing network returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/records?app=demo-lib", nil)
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 for missing network, got %d", rec.Code)
		}
	})
}
