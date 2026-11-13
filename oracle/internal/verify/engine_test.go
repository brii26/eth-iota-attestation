package verify

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/brii26/eth-iota-attestation/oracle/internal/chain"
	"github.com/brii26/eth-iota-attestation/oracle/internal/model"
)

type mockBlobStore struct {
	mu    sync.Mutex
	sboms map[string][]byte
	scans map[string][]byte
}

func newMockBlobStore() *mockBlobStore {
	return &mockBlobStore{
		sboms: make(map[string][]byte),
		scans: make(map[string][]byte),
	}
}

func (m *mockBlobStore) GetSBOM(ctx context.Context, key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, ok := m.sboms[key]
	if !ok {
		return nil, errors.New("sbom not found in s3")
	}
	return data, nil
}

func (m *mockBlobStore) GetScan(ctx context.Context, key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, ok := m.scans[key]
	if !ok {
		return nil, errors.New("scan not found in s3")
	}
	return data, nil
}

type mockArtifactFetcher struct {
	mu        sync.Mutex
	artifacts map[string][]byte // key: app@version
}

func newMockArtifactFetcher() *mockArtifactFetcher {
	return &mockArtifactFetcher{
		artifacts: make(map[string][]byte),
	}
}

func (m *mockArtifactFetcher) FetchArtifact(ctx context.Context, app, version string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s@%s", app, version)
	data, ok := m.artifacts[key]
	if !ok {
		return nil, errors.New("artifact not found in registry")
	}
	return data, nil
}

func setupEngineTest(t *testing.T) (*Engine, *chain.MockAdapter, *mockBlobStore, *mockArtifactFetcher) {
	blobStore := newMockBlobStore()
	fetcher := newMockArtifactFetcher()
	adapter := chain.NewMockAdapter("ethereum-private")

	engine := NewEngine(EngineConfig{
		BlobStore:       blobStore,
		ArtifactFetcher: fetcher,
		MaxGrypeDBAge:   24 * time.Hour,
	})

	return engine, adapter, blobStore, fetcher
}

func TestVerifyEngine(t *testing.T) {
	ctx := context.Background()

	// 1. Success Flow: Valid attestation verification
	t.Run("AC1: Valid attestation returns VALID and stale false", func(t *testing.T) {
		engine, adapter, blobStore, fetcher := setupEngineTest(t)

		app := "demo-lib"
		version := "1.4.2"
		scanID := "scan-12345"
		storageKey := fmt.Sprintf("ethereum-private/%s/%s/%s.json", app, version, scanID)

		artBytes := []byte("tarball-contents-v1.4.2")
		sbomBytes := []byte(`{"bomFormat":"CycloneDX","version":1}`)
		freshTime := time.Now().UTC().Format(time.RFC3339)
		scanBytes := []byte(fmt.Sprintf(`{"descriptor":{"db":{"built":%q}}}`, freshTime))

		artHash, _ := ComputeSHA256(artBytes)
		sbomHash, _ := ComputeSHA256(sbomBytes)
		scanHash, _ := ComputeSHA256(scanBytes)

		fetcher.artifacts[fmt.Sprintf("%s@%s", app, version)] = artBytes
		blobStore.sboms[storageKey] = sbomBytes
		blobStore.scans[storageKey] = scanBytes

		signer := "0x8a3fB9000000000000000000000000000000c19e"
		adapter.SetRecord(chain.ChainRecord{
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

		resp, err := engine.Verify(ctx, adapter, model.VerifyRequest{
			Network: "ethereum-private",
			App:     app,
			Version: version,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Status != model.StatusValid {
			t.Errorf("expected status %s, got %s", model.StatusValid, resp.Status)
		}
		if resp.Stale {
			t.Errorf("expected stale to be false")
		}
		if resp.Checks.Artifact != model.CheckMatch ||
			resp.Checks.SBOM != model.CheckMatch ||
			resp.Checks.Scan != model.CheckMatch ||
			!resp.Checks.SignerTrusted {
			t.Errorf("unexpected checks state: %+v", resp.Checks)
		}
		if resp.RecordID != "42" {
			t.Errorf("expected recordId 42, got %s", resp.RecordID)
		}
	})

	// 2. Tampered artifact in Verdaccio
	t.Run("AC2: Tampered artifact returns ARTIFACT_MISMATCH", func(t *testing.T) {
		engine, adapter, blobStore, fetcher := setupEngineTest(t)

		app := "demo-lib"
		version := "1.4.2"
		scanID := "scan-tampered-art"
		storageKey := fmt.Sprintf("ethereum-private/%s/%s/%s.json", app, version, scanID)

		origArt := []byte("original-artifact")
		tamperedArt := []byte("tampered-malicious-artifact")
		sbomBytes := []byte(`{"bomFormat":"CycloneDX"}`)
		freshTime := time.Now().UTC().Format(time.RFC3339)
		scanBytes := []byte(fmt.Sprintf(`{"descriptor":{"db":{"built":%q}}}`, freshTime))

		artHash, _ := ComputeSHA256(origArt)
		sbomHash, _ := ComputeSHA256(sbomBytes)
		scanHash, _ := ComputeSHA256(scanBytes)

		fetcher.artifacts[fmt.Sprintf("%s@%s", app, version)] = tamperedArt
		blobStore.sboms[storageKey] = sbomBytes
		blobStore.scans[storageKey] = scanBytes

		adapter.SetRecord(chain.ChainRecord{
			RecordID:     "43",
			App:          app,
			Version:      version,
			ScanID:       scanID,
			ArtifactHash: artHash,
			SBOMHash:     sbomHash,
			ScanHash:     scanHash,
			Signer:       "0xSigner",
			Timestamp:    time.Now().UTC(),
		})

		resp, err := engine.Verify(ctx, adapter, model.VerifyRequest{
			Network: "ethereum-private",
			App:     app,
			Version: version,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Status != model.StatusArtifactMismatch {
			t.Errorf("expected status %s, got %s", model.StatusArtifactMismatch, resp.Status)
		}
		if resp.Checks.Artifact != model.CheckMismatch {
			t.Errorf("expected artifact check to be MISMATCH, got %s", resp.Checks.Artifact)
		}
		if resp.Checks.SBOM != model.CheckMatch || resp.Checks.Scan != model.CheckMatch {
			t.Errorf("expected other checks to be MATCH: %+v", resp.Checks)
		}
	})

	// 3. Tampered SBOM or Scan in RustFS
	t.Run("AC3: Tampered SBOM returns SBOM_MISMATCH", func(t *testing.T) {
		engine, adapter, blobStore, fetcher := setupEngineTest(t)

		app := "demo-lib"
		version := "1.4.2"
		scanID := "scan-tampered-sbom"
		storageKey := fmt.Sprintf("ethereum-private/%s/%s/%s.json", app, version, scanID)

		artBytes := []byte("valid-art")
		origSBOM := []byte("orig-sbom")
		tamperedSBOM := []byte("tampered-sbom")
		freshTime := time.Now().UTC().Format(time.RFC3339)
		scanBytes := []byte(fmt.Sprintf(`{"descriptor":{"db":{"built":%q}}}`, freshTime))

		artHash, _ := ComputeSHA256(artBytes)
		origSBOMHash, _ := ComputeSHA256(origSBOM)
		scanHash, _ := ComputeSHA256(scanBytes)

		fetcher.artifacts[fmt.Sprintf("%s@%s", app, version)] = artBytes
		blobStore.sboms[storageKey] = tamperedSBOM
		blobStore.scans[storageKey] = scanBytes

		adapter.SetRecord(chain.ChainRecord{
			RecordID:     "44",
			App:          app,
			Version:      version,
			ScanID:       scanID,
			ArtifactHash: artHash,
			SBOMHash:     origSBOMHash,
			ScanHash:     scanHash,
			Signer:       "0xSigner",
			Timestamp:    time.Now().UTC(),
		})

		resp, err := engine.Verify(ctx, adapter, model.VerifyRequest{
			Network: "ethereum-private",
			App:     app,
			Version: version,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Status != model.StatusSBOMMismatch {
			t.Errorf("expected %s, got %s", model.StatusSBOMMismatch, resp.Status)
		}
		if resp.Checks.SBOM != model.CheckMismatch {
			t.Errorf("expected sbom check to be MISMATCH, got %s", resp.Checks.SBOM)
		}
	})

	t.Run("AC3: Tampered Scan returns SCAN_MISMATCH", func(t *testing.T) {
		engine, adapter, blobStore, fetcher := setupEngineTest(t)

		app := "demo-lib"
		version := "1.4.2"
		scanID := "scan-tampered-scan"
		storageKey := fmt.Sprintf("ethereum-private/%s/%s/%s.json", app, version, scanID)

		artBytes := []byte("valid-art")
		sbomBytes := []byte("valid-sbom")
		origScan := []byte(`{"descriptor":{"db":{"built":"2026-11-13T08:00:00Z"}}}`)
		tamperedScan := []byte(`{"descriptor":{"db":{"built":"2026-11-13T09:00:00Z"}}}`)

		artHash, _ := ComputeSHA256(artBytes)
		sbomHash, _ := ComputeSHA256(sbomBytes)
		origScanHash, _ := ComputeSHA256(origScan)

		fetcher.artifacts[fmt.Sprintf("%s@%s", app, version)] = artBytes
		blobStore.sboms[storageKey] = sbomBytes
		blobStore.scans[storageKey] = tamperedScan

		adapter.SetRecord(chain.ChainRecord{
			RecordID:     "45",
			App:          app,
			Version:      version,
			ScanID:       scanID,
			ArtifactHash: artHash,
			SBOMHash:     sbomHash,
			ScanHash:     origScanHash,
			Signer:       "0xSigner",
			Timestamp:    time.Now().UTC(),
		})

		resp, err := engine.Verify(ctx, adapter, model.VerifyRequest{
			Network: "ethereum-private",
			App:     app,
			Version: version,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Status != model.StatusScanMismatch {
			t.Errorf("expected %s, got %s", model.StatusScanMismatch, resp.Status)
		}
		if resp.Checks.Scan != model.CheckMismatch {
			t.Errorf("expected scan check to be MISMATCH, got %s", resp.Checks.Scan)
		}
	})

	// 4. Revoke oracle
	t.Run("AC4: Revoked oracle returns UNTRUSTED_SIGNER", func(t *testing.T) {
		engine, adapter, blobStore, fetcher := setupEngineTest(t)

		app := "demo-lib"
		version := "1.4.2"
		scanID := "scan-revoked"
		storageKey := fmt.Sprintf("ethereum-private/%s/%s/%s.json", app, version, scanID)

		artBytes := []byte("valid-art")
		sbomBytes := []byte("valid-sbom")
		freshTime := time.Now().UTC().Format(time.RFC3339)
		scanBytes := []byte(fmt.Sprintf(`{"descriptor":{"db":{"built":%q}}}`, freshTime))

		artHash, _ := ComputeSHA256(artBytes)
		sbomHash, _ := ComputeSHA256(sbomBytes)
		scanHash, _ := ComputeSHA256(scanBytes)

		fetcher.artifacts[fmt.Sprintf("%s@%s", app, version)] = artBytes
		blobStore.sboms[storageKey] = sbomBytes
		blobStore.scans[storageKey] = scanBytes

		revokedSigner := "0xRevokedOracleSigner000000000000000000000"
		adapter.SetRecord(chain.ChainRecord{
			RecordID:     "46",
			App:          app,
			Version:      version,
			ScanID:       scanID,
			ArtifactHash: artHash,
			SBOMHash:     sbomHash,
			ScanHash:     scanHash,
			Signer:       revokedSigner,
			Timestamp:    time.Now().UTC(),
		})

		// Revoke signer in adapter
		adapter.SetSignerAuthorized(revokedSigner, false)

		resp, err := engine.Verify(ctx, adapter, model.VerifyRequest{
			Network: "ethereum-private",
			App:     app,
			Version: version,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Status != model.StatusUntrustedSigner {
			t.Errorf("expected status %s, got %s", model.StatusUntrustedSigner, resp.Status)
		}
		if resp.Checks.SignerTrusted {
			t.Errorf("expected signerTrusted to be false")
		}
		if resp.Checks.Artifact != model.CheckMatch || resp.Checks.SBOM != model.CheckMatch || resp.Checks.Scan != model.CheckMatch {
			t.Errorf("expected file checks to be MATCH: %+v", resp.Checks)
		}
	})

	// 5. Non-existent app and missing file
	t.Run("AC5: Non-existent app returns RECORD_NOT_FOUND", func(t *testing.T) {
		engine, adapter, _, _ := setupEngineTest(t)

		resp, err := engine.Verify(ctx, adapter, model.VerifyRequest{
			Network: "ethereum-private",
			App:     "missing-app",
			Version: "1.0.0",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Status != model.StatusRecordNotFound {
			t.Errorf("expected %s, got %s", model.StatusRecordNotFound, resp.Status)
		}
		if resp.Checks.Artifact != model.CheckUnavailable ||
			resp.Checks.SBOM != model.CheckUnavailable ||
			resp.Checks.Scan != model.CheckUnavailable {
			t.Errorf("expected checks to be UNAVAILABLE: %+v", resp.Checks)
		}
	})

	t.Run("AC5: Missing file returns FILE_UNAVAILABLE", func(t *testing.T) {
		engine, adapter, blobStore, fetcher := setupEngineTest(t)

		app := "demo-lib"
		version := "1.4.2"
		scanID := "scan-missing-file"
		storageKey := fmt.Sprintf("ethereum-private/%s/%s/%s.json", app, version, scanID)

		artBytes := []byte("valid-art")
		freshTime := time.Now().UTC().Format(time.RFC3339)
		scanBytes := []byte(fmt.Sprintf(`{"descriptor":{"db":{"built":%q}}}`, freshTime))

		artHash, _ := ComputeSHA256(artBytes)
		sbomHash, _ := ComputeSHA256([]byte("missing-sbom"))
		scanHash, _ := ComputeSHA256(scanBytes)

		fetcher.artifacts[fmt.Sprintf("%s@%s", app, version)] = artBytes
		// sbomBytes omitted from blobStore -> simulated missing file
		blobStore.scans[storageKey] = scanBytes

		adapter.SetRecord(chain.ChainRecord{
			RecordID:     "47",
			App:          app,
			Version:      version,
			ScanID:       scanID,
			ArtifactHash: artHash,
			SBOMHash:     sbomHash,
			ScanHash:     scanHash,
			Signer:       "0xSigner",
			Timestamp:    time.Now().UTC(),
		})

		resp, err := engine.Verify(ctx, adapter, model.VerifyRequest{
			Network: "ethereum-private",
			App:     app,
			Version: version,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Status != model.StatusFileUnavailable {
			t.Errorf("expected %s, got %s", model.StatusFileUnavailable, resp.Status)
		}
		if resp.Checks.SBOM != model.CheckUnavailable {
			t.Errorf("expected SBOM check to be UNAVAILABLE, got %s", resp.Checks.SBOM)
		}
	})

	// 6. Stale Grype DB
	t.Run("AC6: Stale Grype DB returns VALID with stale true", func(t *testing.T) {
		engine, adapter, blobStore, fetcher := setupEngineTest(t)

		app := "demo-lib"
		version := "1.4.2"
		scanID := "scan-stale-grype"
		storageKey := fmt.Sprintf("ethereum-private/%s/%s/%s.json", app, version, scanID)

		artBytes := []byte("valid-art")
		sbomBytes := []byte("valid-sbom")
		staleTime := time.Now().UTC().Add(-72 * time.Hour).Format(time.RFC3339)
		staleScanBytes := []byte(fmt.Sprintf(`{"descriptor":{"db":{"built":%q}}}`, staleTime))

		artHash, _ := ComputeSHA256(artBytes)
		sbomHash, _ := ComputeSHA256(sbomBytes)
		scanHash, _ := ComputeSHA256(staleScanBytes)

		fetcher.artifacts[fmt.Sprintf("%s@%s", app, version)] = artBytes
		blobStore.sboms[storageKey] = sbomBytes
		blobStore.scans[storageKey] = staleScanBytes

		adapter.SetRecord(chain.ChainRecord{
			RecordID:     "48",
			App:          app,
			Version:      version,
			ScanID:       scanID,
			ArtifactHash: artHash,
			SBOMHash:     sbomHash,
			ScanHash:     scanHash,
			Signer:       "0xSigner",
			Timestamp:    time.Now().UTC(),
		})

		resp, err := engine.Verify(ctx, adapter, model.VerifyRequest{
			Network: "ethereum-private",
			App:     app,
			Version: version,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Status != model.StatusValid {
			t.Errorf("expected status %s, got %s", model.StatusValid, resp.Status)
		}
		if !resp.Stale {
			t.Errorf("expected stale to be true for stale grype DB")
		}
		if resp.Checks.Artifact != model.CheckMatch || resp.Checks.SBOM != model.CheckMatch || resp.Checks.Scan != model.CheckMatch {
			t.Errorf("expected all checks to match: %+v", resp.Checks)
		}
	})

	// 7. Verify with explicit RecordID
	t.Run("Verify with explicit RecordID", func(t *testing.T) {
		engine, adapter, blobStore, fetcher := setupEngineTest(t)

		app := "demo-lib"
		version := "1.4.2"
		scanID := "scan-record-id"
		storageKey := fmt.Sprintf("ethereum-private/%s/%s/%s.json", app, version, scanID)

		artBytes := []byte("valid-art")
		sbomBytes := []byte("valid-sbom")
		freshTime := time.Now().UTC().Format(time.RFC3339)
		scanBytes := []byte(fmt.Sprintf(`{"descriptor":{"db":{"built":%q}}}`, freshTime))

		artHash, _ := ComputeSHA256(artBytes)
		sbomHash, _ := ComputeSHA256(sbomBytes)
		scanHash, _ := ComputeSHA256(scanBytes)

		fetcher.artifacts[fmt.Sprintf("%s@%s", app, version)] = artBytes
		blobStore.sboms[storageKey] = sbomBytes
		blobStore.scans[storageKey] = scanBytes

		adapter.SetRecord(chain.ChainRecord{
			RecordID:     "999",
			App:          app,
			Version:      version,
			ScanID:       scanID,
			ArtifactHash: artHash,
			SBOMHash:     sbomHash,
			ScanHash:     scanHash,
			Signer:       "0xSigner",
			Timestamp:    time.Now().UTC(),
		})

		resp, err := engine.Verify(ctx, adapter, model.VerifyRequest{
			Network:  "ethereum-private",
			App:      app,
			Version:  version,
			RecordID: "999",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Status != model.StatusValid || resp.RecordID != "999" {
			t.Errorf("expected valid status with recordId 999, got status=%s, id=%s", resp.Status, resp.RecordID)
		}
	})
}
