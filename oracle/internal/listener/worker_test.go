package listener

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/brii26/eth-iota-attestation/oracle/internal/chain"
	"github.com/brii26/eth-iota-attestation/oracle/internal/store"
	"github.com/brii26/eth-iota-attestation/oracle/internal/verify"
)

type mockBlobStore struct {
	mu    sync.Mutex
	sboms map[string][]byte
	scans map[string][]byte
}

func (m *mockBlobStore) GetSBOM(ctx context.Context, key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, ok := m.sboms[key]
	if !ok {
		return nil, errors.New("sbom not found in test blob store")
	}
	return data, nil
}

func (m *mockBlobStore) GetScan(ctx context.Context, key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, ok := m.scans[key]
	if !ok {
		return nil, errors.New("scan not found in test blob store")
	}
	return data, nil
}

type mockArtifactFetcher struct {
	mu        sync.Mutex
	artifacts map[string][]byte
}

func (m *mockArtifactFetcher) FetchArtifact(ctx context.Context, app, version string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, ok := m.artifacts[fmt.Sprintf("%s@%s", app, version)]
	if !ok {
		return nil, errors.New("artifact not found in registry")
	}
	return data, nil
}

func setupWorkerTest(t *testing.T) (*Worker, *chain.MockAdapter, *store.SQLiteIndex, *mockBlobStore, *mockArtifactFetcher) {
	tmpDir := t.TempDir()
	idx, err := store.NewSQLiteIndex(filepath.Join(tmpDir, "test-worker.sqlite"))
	if err != nil {
		t.Fatalf("failed to init sqlite: %v", err)
	}

	adapter := chain.NewMockAdapter("ethereum-private")
	blobStore := &mockBlobStore{
		sboms: make(map[string][]byte),
		scans: make(map[string][]byte),
	}
	fetcher := &mockArtifactFetcher{
		artifacts: make(map[string][]byte),
	}

	engine := verify.NewEngine(verify.EngineConfig{
		BlobStore:       blobStore,
		ArtifactFetcher: fetcher,
		MaxGrypeDBAge:   24 * time.Hour,
	})

	worker := NewWorker(Config{
		Adapter:      adapter,
		Engine:       engine,
		Index:        idx,
		Network:      "ethereum-private",
		PollInterval: 10 * time.Millisecond,
		RetryBackoff: 5 * time.Millisecond,
		MaxRetries:   2,
	})

	return worker, adapter, idx, blobStore, fetcher
}

func TestVerificationListener(t *testing.T) {
	ctx := context.Background()

	// 1. AC1: requestVerification answered and displays correct result
	t.Run("AC1: Valid verification request answered on-chain", func(t *testing.T) {
		worker, adapter, idx, blobStore, fetcher := setupWorkerTest(t)
		defer idx.Close()

		app := "demo-lib"
		version := "1.4.2"
		scanID := "scan-req-1"
		storageKey := fmt.Sprintf("ethereum-private/%s/%s/%s.json", app, version, scanID)

		artBytes := []byte("valid-art-bytes")
		sbomBytes := []byte(`{"bomFormat":"CycloneDX"}`)
		freshTime := time.Now().UTC().Format(time.RFC3339)
		scanBytes := []byte(fmt.Sprintf(`{"descriptor":{"db":{"built":%q}}}`, freshTime))

		artHash, _ := verify.ComputeSHA256(artBytes)
		sbomHash, _ := verify.ComputeSHA256(sbomBytes)
		scanHash, _ := verify.ComputeSHA256(scanBytes)

		fetcher.artifacts[fmt.Sprintf("%s@%s", app, version)] = artBytes
		blobStore.sboms[storageKey] = sbomBytes
		blobStore.scans[storageKey] = scanBytes

		adapter.SetRecord(chain.ChainRecord{
			RecordID:     "100",
			App:          app,
			Version:      version,
			ScanID:       scanID,
			ArtifactHash: artHash,
			SBOMHash:     sbomHash,
			ScanHash:     scanHash,
			Signer:       "0xOracle",
			Timestamp:    time.Now().UTC(),
		})

		// User requests verification on-chain at block 10
		adapter.AddVerificationRequest(big.NewInt(1), big.NewInt(100), "0xUserAddress", 10)

		processed, err := worker.PollOnce(ctx)
		if err != nil {
			t.Fatalf("unexpected poll error: %v", err)
		}
		if processed != 1 {
			t.Fatalf("expected 1 processed event, got %d", processed)
		}

		res, ok := adapter.GetFulfilledVerification(big.NewInt(1))
		if !ok {
			t.Fatalf("expected request 1 to be fulfilled on adapter")
		}
		// Status 1 is Valid, stale is false
		if res.Status != 1 || res.Stale {
			t.Errorf("expected status 1 (Valid) and stale false, got status=%d stale=%v", res.Status, res.Stale)
		}

		// Verify recorded in SQLite
		fulfilled, err := idx.IsVerificationFulfilled(ctx, "ethereum-private", "1")
		if err != nil || !fulfilled {
			t.Errorf("expected sqlite to record request 1 fulfilled, err=%v", err)
		}
	})

	// 2. AC2: Modified file produces status mismatch on-chain
	t.Run("AC2: Tampered artifact produces ARTIFACT_MISMATCH status on-chain", func(t *testing.T) {
		worker, adapter, idx, blobStore, fetcher := setupWorkerTest(t)
		defer idx.Close()

		app := "demo-lib"
		version := "1.4.2"
		scanID := "scan-req-2"
		storageKey := fmt.Sprintf("ethereum-private/%s/%s/%s.json", app, version, scanID)

		origArt := []byte("original-artifact")
		tamperedArt := []byte("tampered-malicious-artifact")
		sbomBytes := []byte(`{"bomFormat":"CycloneDX"}`)
		freshTime := time.Now().UTC().Format(time.RFC3339)
		scanBytes := []byte(fmt.Sprintf(`{"descriptor":{"db":{"built":%q}}}`, freshTime))

		artHash, _ := verify.ComputeSHA256(origArt)
		sbomHash, _ := verify.ComputeSHA256(sbomBytes)
		scanHash, _ := verify.ComputeSHA256(scanBytes)

		fetcher.artifacts[fmt.Sprintf("%s@%s", app, version)] = tamperedArt // Registry has tampered art
		blobStore.sboms[storageKey] = sbomBytes
		blobStore.scans[storageKey] = scanBytes

		adapter.SetRecord(chain.ChainRecord{
			RecordID:     "101",
			App:          app,
			Version:      version,
			ScanID:       scanID,
			ArtifactHash: artHash,
			SBOMHash:     sbomHash,
			ScanHash:     scanHash,
			Signer:       "0xOracle",
			Timestamp:    time.Now().UTC(),
		})

		adapter.AddVerificationRequest(big.NewInt(2), big.NewInt(101), "0xUserAddress", 20)

		processed, err := worker.PollOnce(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if processed != 1 {
			t.Fatalf("expected 1 processed event, got %d", processed)
		}

		res, ok := adapter.GetFulfilledVerification(big.NewInt(2))
		if !ok {
			t.Fatalf("expected request 2 to be fulfilled")
		}
		// Status 2 is ArtifactMismatch
		if res.Status != 2 {
			t.Errorf("expected status 2 (ArtifactMismatch), got %d", res.Status)
		}
	})

	// 3. AC3: Oracle restart does not answer twice and catches up missed requests
	t.Run("AC3: Oracle restart does not answer twice and catches up missed requests", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "persistent.sqlite")

		idx1, err := store.NewSQLiteIndex(dbPath)
		if err != nil {
			t.Fatalf("init sqlite1: %v", err)
		}

		adapter := chain.NewMockAdapter("ethereum-private")
		blobStore := &mockBlobStore{
			sboms: make(map[string][]byte),
			scans: make(map[string][]byte),
		}
		fetcher := &mockArtifactFetcher{
			artifacts: make(map[string][]byte),
		}
		engine := verify.NewEngine(verify.EngineConfig{
			BlobStore:       blobStore,
			ArtifactFetcher: fetcher,
		})

		worker1 := NewWorker(Config{
			Adapter: adapter,
			Engine:  engine,
			Index:   idx1,
			Network: "ethereum-private",
		})

		app := "demo-lib"
		version := "1.4.2"
		scanID := "scan-req-3"
		storageKey := fmt.Sprintf("ethereum-private/%s/%s/%s.json", app, version, scanID)

		artBytes := []byte("valid-art")
		sbomBytes := []byte("valid-sbom")
		freshTime := time.Now().UTC().Format(time.RFC3339)
		scanBytes := []byte(fmt.Sprintf(`{"descriptor":{"db":{"built":%q}}}`, freshTime))

		artHash, _ := verify.ComputeSHA256(artBytes)
		sbomHash, _ := verify.ComputeSHA256(sbomBytes)
		scanHash, _ := verify.ComputeSHA256(scanBytes)

		fetcher.artifacts[fmt.Sprintf("%s@%s", app, version)] = artBytes
		blobStore.sboms[storageKey] = sbomBytes
		blobStore.scans[storageKey] = scanBytes

		adapter.SetRecord(chain.ChainRecord{
			RecordID:     "102",
			App:          app,
			Version:      version,
			ScanID:       scanID,
			ArtifactHash: artHash,
			SBOMHash:     sbomHash,
			ScanHash:     scanHash,
			Signer:       "0xOracle",
			Timestamp:    time.Now().UTC(),
		})

		// Event 1 at block 10
		adapter.AddVerificationRequest(big.NewInt(3), big.NewInt(102), "0xUser", 10)
		p1, err := worker1.PollOnce(ctx)
		if err != nil || p1 != 1 {
			t.Fatalf("worker1 poll: p1=%d err=%v", p1, err)
		}

		// Simulate oracle stopping/crashing: close idx1
		idx1.Close()

		// While oracle is down, event 2 happens at block 25
		adapter.AddVerificationRequest(big.NewInt(4), big.NewInt(102), "0xUser", 25)

		// Restart oracle: open new SQLite connection to the same file
		idx2, err := store.NewSQLiteIndex(dbPath)
		if err != nil {
			t.Fatalf("init sqlite2: %v", err)
		}
		defer idx2.Close()

		worker2 := NewWorker(Config{
			Adapter: adapter,
			Engine:  engine,
			Index:   idx2,
			Network: "ethereum-private",
		})

		// Worker2 polls: it should process only the new request (event 4) and NOT re-process event 3
		p2, err := worker2.PollOnce(ctx)
		if err != nil {
			t.Fatalf("worker2 poll error: %v", err)
		}
		if p2 != 1 {
			t.Fatalf("expected exactly 1 new event processed upon restart, got %d", p2)
		}

		// Ensure request 4 was fulfilled
		if _, ok := adapter.GetFulfilledVerification(big.NewInt(4)); !ok {
			t.Errorf("expected request 4 to be fulfilled after restart")
		}

		// Poll again immediately: should process 0 events (no duplicate fulfillments)
		p3, err := worker2.PollOnce(ctx)
		if err != nil || p3 != 0 {
			t.Fatalf("expected 0 events on second poll, got %d, err=%v", p3, err)
		}
	})

	// 4. Retry on temporary failure
	t.Run("Retry on temporary fulfillment failure", func(t *testing.T) {
		worker, adapter, idx, blobStore, fetcher := setupWorkerTest(t)
		defer idx.Close()

		app := "demo-lib"
		version := "1.4.2"
		scanID := "scan-req-retry"
		storageKey := fmt.Sprintf("ethereum-private/%s/%s/%s.json", app, version, scanID)

		artBytes := []byte("valid-art")
		sbomBytes := []byte("valid-sbom")
		freshTime := time.Now().UTC().Format(time.RFC3339)
		scanBytes := []byte(fmt.Sprintf(`{"descriptor":{"db":{"built":%q}}}`, freshTime))

		artHash, _ := verify.ComputeSHA256(artBytes)
		sbomHash, _ := verify.ComputeSHA256(sbomBytes)
		scanHash, _ := verify.ComputeSHA256(scanBytes)

		fetcher.artifacts[fmt.Sprintf("%s@%s", app, version)] = artBytes
		blobStore.sboms[storageKey] = sbomBytes
		blobStore.scans[storageKey] = scanBytes

		adapter.SetRecord(chain.ChainRecord{
			RecordID:     "103",
			App:          app,
			Version:      version,
			ScanID:       scanID,
			ArtifactHash: artHash,
			SBOMHash:     sbomHash,
			ScanHash:     scanHash,
			Signer:       "0xOracle",
			Timestamp:    time.Now().UTC(),
		})

		adapter.AddVerificationRequest(big.NewInt(5), big.NewInt(103), "0xUser", 30)

		// Set adapter to fail
		adapter.SetFail(true)

		_, err := worker.PollOnce(ctx)
		if err == nil {
			t.Fatalf("expected poll to fail when adapter fails")
		}

		// Recovery: un-set failure
		adapter.SetFail(false)

		// Re-poll: should successfully retry and fulfill request 5
		p, err := worker.PollOnce(ctx)
		if err != nil || p != 1 {
			t.Fatalf("expected successful retry, got p=%d err=%v", p, err)
		}

		if _, ok := adapter.GetFulfilledVerification(big.NewInt(5)); !ok {
			t.Errorf("expected request 5 to be fulfilled after retry")
		}
	})
}
