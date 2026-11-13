package chain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Record is the normalized off-chain attestation record to be written to a target chain.
type Record struct {
	Network      string
	App          string
	Version      string
	ScanID       string
	ArtifactHash [32]byte
	SBOMHash     [32]byte
	ScanHash     [32]byte
}

// Receipt is returned after successfully writing a record to a chain adapter.
type Receipt struct {
	TxRef      string
	RecordID   string
	IncludedAt time.Time
}

// Adapter defines the interface for chain-specific attestation writers.
type Adapter interface {
	WriteRecord(ctx context.Context, rec Record) (*Receipt, error)
}

// MockAdapter provides an in-memory chain adapter for tests and standalone oracle runs.
type MockAdapter struct {
	mu      sync.Mutex
	records []Record
	fail    bool
}

// NewMockAdapter constructs a new MockAdapter.
func NewMockAdapter() *MockAdapter {
	return &MockAdapter{records: make([]Record, 0)}
}

// SetFail configures the mock to fail subsequent calls.
func (m *MockAdapter) SetFail(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fail = fail
}

// WriteRecord stores the record in memory and returns a deterministic simulated receipt.
func (m *MockAdapter) WriteRecord(ctx context.Context, rec Record) (*Receipt, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.fail {
		return nil, fmt.Errorf("mock adapter write failure")
	}

	m.records = append(m.records, rec)
	txHash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s", rec.Network, rec.ScanID, time.Now().UTC().String())))
	recID := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s", rec.App, rec.Version, rec.ScanID)))

	return &Receipt{
		TxRef:      "0x" + hex.EncodeToString(txHash[:]),
		RecordID:   "0x" + hex.EncodeToString(recID[:]),
		IncludedAt: time.Now().UTC(),
	}, nil
}

// Records returns a slice of all recorded records.
func (m *MockAdapter) Records() []Record {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make([]Record, len(m.records))
	copy(copied, m.records)
	return copied
}
