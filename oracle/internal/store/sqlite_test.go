package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestSQLiteIndex(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.sqlite")

	idx, err := NewSQLiteIndex(dbPath)
	if err != nil {
		t.Fatalf("failed to create sqlite index: %v", err)
	}
	defer idx.Close()

	rec := AttestationRecord{
		Network:      "ethereum-private",
		App:          "demo-app",
		Version:      "1.0.0",
		ScanID:       "0f8c2d3e-6b1a-4c55-9e1f-2a7b8c9d0e11",
		ArtifactHash: "hash-art",
		SBOMHash:     "hash-sbom",
		ScanHash:     "hash-scan",
		Status:       "submitted",
		SubmittedAt:  time.Now().UTC(),
	}

	// 1. Should not exist initially
	exists, err := idx.Exists(ctx, rec.Network, rec.ScanID)
	if err != nil {
		t.Fatalf("unexpected error checking existence: %v", err)
	}
	if exists {
		t.Fatalf("expected record to not exist initially")
	}

	// 2. Save
	if err := idx.Save(ctx, rec); err != nil {
		t.Fatalf("failed to save record: %v", err)
	}

	// 3. Duplicate should fail unique constraint
	if err := idx.Save(ctx, rec); err == nil {
		t.Fatalf("expected duplicate insert to fail, got nil")
	}

	// 4. Exists should be true
	exists, err = idx.Exists(ctx, rec.Network, rec.ScanID)
	if err != nil || !exists {
		t.Fatalf("expected record to exist, err=%v", err)
	}

	// 5. Query status
	st, err := idx.GetStatus(ctx, rec.Network, rec.ScanID)
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}
	if st == nil || st.Status != "submitted" {
		t.Fatalf("unexpected status: %+v", st)
	}

	// 6. Update status
	now := time.Now().UTC()
	err = idx.UpdateStatus(ctx, rec.Network, rec.ScanID, "included", "0xabc", "0x123", &now)
	if err != nil {
		t.Fatalf("failed to update status: %v", err)
	}

	st, err = idx.GetStatus(ctx, rec.Network, rec.ScanID)
	if err != nil {
		t.Fatalf("failed to get updated status: %v", err)
	}
	if st.Status != "included" || st.TxRef != "0xabc" || st.RecordID != "0x123" {
		t.Fatalf("status update mismatch: %+v", st)
	}

	// 7. Sync state (last processed block)
	block, err := idx.GetLastProcessedBlock(ctx, rec.Network)
	if err != nil || block != 0 {
		t.Fatalf("expected initial block 0, got %d, err=%v", block, err)
	}

	if err := idx.SetLastProcessedBlock(ctx, rec.Network, 100); err != nil {
		t.Fatalf("failed to set last processed block: %v", err)
	}

	block, err = idx.GetLastProcessedBlock(ctx, rec.Network)
	if err != nil || block != 100 {
		t.Fatalf("expected block 100, got %d, err=%v", block, err)
	}

	// 8. Fulfilled verifications
	fulfilled, err := idx.IsVerificationFulfilled(ctx, rec.Network, "req-1")
	if err != nil || fulfilled {
		t.Fatalf("expected req-1 not fulfilled, got %v, err=%v", fulfilled, err)
	}

	if err := idx.RecordFulfilledVerification(ctx, rec.Network, "req-1", "rec-42", 1, false, "0xtx"); err != nil {
		t.Fatalf("failed to record fulfilled verification: %v", err)
	}

	fulfilled, err = idx.IsVerificationFulfilled(ctx, rec.Network, "req-1")
	if err != nil || !fulfilled {
		t.Fatalf("expected req-1 fulfilled, got %v, err=%v", fulfilled, err)
	}
}
