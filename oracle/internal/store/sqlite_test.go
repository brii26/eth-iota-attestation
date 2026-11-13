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
}
