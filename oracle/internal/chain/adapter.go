package chain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// ErrRecordNotFound is returned when an on-chain record cannot be found.
var ErrRecordNotFound = errors.New("record not found on chain")

// ErrNetworkUnregistered is returned when a requested network is not configured.
var ErrNetworkUnregistered = errors.New("network is not registered")

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
	TxRef       string
	RecordID    string
	BlockNumber uint64
	GasUsed     uint64
	TxSize      uint64
	SubmittedAt time.Time
	IncludedAt  time.Time
}

// Query specifies parameters for reading on-chain records.
type Query struct {
	App      string
	Version  string
	RecordID *big.Int
}

// ChainRecord represents a record retrieved directly from a blockchain.
type ChainRecord struct {
	RecordID     string
	App          string
	Version      string
	ScanID       string
	ArtifactHash [32]byte
	SBOMHash     [32]byte
	ScanHash     [32]byte
	Signer       string
	Timestamp    time.Time
}

// Adapter defines the common interface for chain-specific attestation writers and readers.
type Adapter interface {
	WriteRecord(ctx context.Context, rec Record) (*Receipt, error)
	ReadRecord(ctx context.Context, query Query) (*ChainRecord, error)
	IsSignerAuthorized(ctx context.Context, signer string) (bool, error)
	Name() string
}

// MockAdapter provides an in-memory chain adapter for tests and standalone oracle runs.
type MockAdapter struct {
	name             string
	mu               sync.Mutex
	records          map[string]ChainRecord // key: app@version
	receipts         map[string]Receipt     // key: network:scanId
	fail             bool
	untrustedSigners map[string]bool
}

// NewMockAdapter constructs a new MockAdapter with an optional name (defaults to "ethereum-private").
func NewMockAdapter(name ...string) *MockAdapter {
	adapterName := "ethereum-private"
	if len(name) > 0 && name[0] != "" {
		adapterName = name[0]
	}
	return &MockAdapter{
		name:             adapterName,
		records:          make(map[string]ChainRecord),
		receipts:         make(map[string]Receipt),
		untrustedSigners: make(map[string]bool),
	}
}

// Name returns the identifier of the mock adapter.
func (m *MockAdapter) Name() string {
	if m.name != "" {
		return m.name
	}
	return "mock"
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

	key := fmt.Sprintf("%s:%s", rec.Network, rec.ScanID)
	// Idempotency: return existing receipt if already written
	if existing, ok := m.receipts[key]; ok {
		return &existing, nil
	}

	txHash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s", rec.Network, rec.ScanID, time.Now().UTC().String())))
	recID := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s", rec.App, rec.Version, rec.ScanID)))

	now := time.Now().UTC()
	receipt := Receipt{
		TxRef:       "0x" + hex.EncodeToString(txHash[:]),
		RecordID:    "0x" + hex.EncodeToString(recID[:]),
		BlockNumber: 42,
		GasUsed:     125000,
		TxSize:      384,
		SubmittedAt: now,
		IncludedAt:  now,
	}

	m.receipts[key] = receipt
	m.records[fmt.Sprintf("%s@%s", rec.App, rec.Version)] = ChainRecord{
		RecordID:     receipt.RecordID,
		App:          rec.App,
		Version:      rec.Version,
		ScanID:       rec.ScanID,
		ArtifactHash: rec.ArtifactHash,
		SBOMHash:     rec.SBOMHash,
		ScanHash:     rec.ScanHash,
		Signer:       "0xMockOracleSignerAddress0000000000000000",
		Timestamp:    now,
	}

	return &receipt, nil
}

// ReadRecord looks up a record by recordId or app and version in the mock store.
func (m *MockAdapter) ReadRecord(ctx context.Context, query Query) (*ChainRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if query.RecordID != nil {
		idDec := query.RecordID.String()
		idHex := "0x" + query.RecordID.Text(16)
		for _, rec := range m.records {
			if rec.RecordID == idDec || rec.RecordID == idHex {
				return &rec, nil
			}
		}
		return nil, ErrRecordNotFound
	}

	key := fmt.Sprintf("%s@%s", query.App, query.Version)
	rec, ok := m.records[key]
	if !ok {
		return nil, ErrRecordNotFound
	}
	return &rec, nil
}

// SetRecord manually sets or overrides a record in the mock store.
func (m *MockAdapter) SetRecord(rec ChainRecord) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records[fmt.Sprintf("%s@%s", rec.App, rec.Version)] = rec
}

// SetSignerAuthorized configures whether a signer is considered an authorized oracle in the mock.
func (m *MockAdapter) SetSignerAuthorized(signer string, authorized bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.untrustedSigners[signer] = !authorized
}

// IsSignerAuthorized checks if the signer address is currently authorized.
func (m *MockAdapter) IsSignerAuthorized(ctx context.Context, signer string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.untrustedSigners[signer] {
		return false, nil
	}
	return true, nil
}

// Records returns a slice of all recorded records in the mock.
func (m *MockAdapter) Records() []ChainRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make([]ChainRecord, 0, len(m.records))
	for _, r := range m.records {
		copied = append(copied, r)
	}
	return copied
}
