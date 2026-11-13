package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/brii26/eth-iota-attestation/oracle/internal/model"
)

// AttestationRecord holds an entry to be written to the SQLite index.
type AttestationRecord struct {
	Network      string
	App          string
	Version      string
	ScanID       string
	ArtifactHash string
	SBOMHash     string
	ScanHash     string
	Status       string
	TxRef        string
	RecordID     string
	SubmittedAt  time.Time
	IncludedAt   *time.Time
}

// SQLiteIndex manages the rebuildable off-chain cache and status index.
type SQLiteIndex struct {
	db *sql.DB
}

// NewSQLiteIndex opens the SQLite database and executes table migrations.
func NewSQLiteIndex(dbPath string) (*SQLiteIndex, error) {
	if dbPath == "" {
		dbPath = "attestations.sqlite"
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	// Enable WAL mode for high concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set sqlite pragmas: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS attestations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		network TEXT NOT NULL,
		app TEXT NOT NULL,
		version TEXT NOT NULL,
		scan_id TEXT NOT NULL,
		artifact_hash TEXT NOT NULL,
		sbom_hash TEXT NOT NULL,
		scan_hash TEXT NOT NULL,
		status TEXT NOT NULL,
		tx_ref TEXT,
		record_id TEXT,
		submitted_at DATETIME NOT NULL,
		included_at DATETIME,
		UNIQUE(network, scan_id)
	);
	CREATE INDEX IF NOT EXISTS idx_attestations_lookup ON attestations(network, scan_id);
	`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("init sqlite schema: %w", err)
	}

	return &SQLiteIndex{db: db}, nil
}

// Exists checks if a record for the specified network and scanId already exists.
func (s *SQLiteIndex) Exists(ctx context.Context, network, scanID string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM attestations WHERE network = ? AND scan_id = ?", network, scanID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("query attestation existence: %w", err)
	}
	return count > 0, nil
}

// Save inserts an attestation record in the database.
func (s *SQLiteIndex) Save(ctx context.Context, rec AttestationRecord) error {
	query := `
	INSERT INTO attestations (
		network, app, version, scan_id, artifact_hash, sbom_hash, scan_hash,
		status, tx_ref, record_id, submitted_at, included_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.ExecContext(ctx, query,
		rec.Network, rec.App, rec.Version, rec.ScanID, rec.ArtifactHash, rec.SBOMHash, rec.ScanHash,
		rec.Status, rec.TxRef, rec.RecordID, rec.SubmittedAt, rec.IncludedAt,
	)
	if err != nil {
		return fmt.Errorf("insert attestation record: %w", err)
	}
	return nil
}

// UpdateStatus updates the transaction reference, status, and includedAt time.
func (s *SQLiteIndex) UpdateStatus(ctx context.Context, network, scanID, status, txRef, recordID string, includedAt *time.Time) error {
	query := `
	UPDATE attestations
	SET status = ?, tx_ref = ?, record_id = ?, included_at = ?
	WHERE network = ? AND scan_id = ?
	`
	res, err := s.db.ExecContext(ctx, query, status, txRef, recordID, includedAt, network, scanID)
	if err != nil {
		return fmt.Errorf("update attestation status: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if rows == 0 {
		return errors.New("no record found to update")
	}
	return nil
}

// GetStatus retrieves the status and receipt info for an attestation.
func (s *SQLiteIndex) GetStatus(ctx context.Context, network, scanID string) (*model.AttestationStatus, error) {
	query := `
	SELECT network, scan_id, status, COALESCE(tx_ref, ''), COALESCE(record_id, ''), submitted_at, included_at
	FROM attestations
	WHERE network = ? AND scan_id = ?
	`
	row := s.db.QueryRowContext(ctx, query, network, scanID)

	var st model.AttestationStatus
	var includedAt sql.NullTime
	err := row.Scan(&st.Network, &st.ScanID, &st.Status, &st.TxRef, &st.RecordID, &st.SubmittedAt, &includedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, fmt.Errorf("scan attestation status: %w", err)
	}

	if includedAt.Valid {
		st.IncludedAt = &includedAt.Time
	}

	return &st, nil
}

// Close closes the underlying SQLite database connection.
func (s *SQLiteIndex) Close() error {
	return s.db.Close()
}
