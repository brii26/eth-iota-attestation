package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type IOTATransaction struct {
	Digest    string `json:"digest"`
	Bytes     string `json:"bytes"`
	Signature string `json:"signature"`
}

func (s *SQLiteIndex) IOTATransaction(ctx context.Context, network, scanID string) (*IOTATransaction, error) {
	var tx IOTATransaction
	err := s.db.QueryRowContext(ctx, `
		SELECT digest, tx_bytes, signature FROM iota_transactions WHERE network = ? AND scan_id = ?
	`, network, scanID).Scan(&tx.Digest, &tx.Bytes, &tx.Signature)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &tx, err
}

func (s *SQLiteIndex) SaveIOTATransaction(ctx context.Context, network, scanID string, tx IOTATransaction) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO iota_transactions (network, scan_id, digest, tx_bytes, signature)
		VALUES (?, ?, ?, ?, ?) ON CONFLICT(network, scan_id) DO NOTHING
	`, network, scanID, tx.Digest, tx.Bytes, tx.Signature)
	return err
}

func (s *SQLiteIndex) PendingIOTATransactions(ctx context.Context, network string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT a.scan_id FROM attestations a JOIN iota_transactions t
		ON a.network = t.network AND a.scan_id = t.scan_id
		WHERE a.network = ? AND a.status = 'pending'
		ORDER BY a.id
	`, network)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scanIDs []string
	for rows.Next() {
		var scanID string
		if err := rows.Scan(&scanID); err != nil {
			return nil, err
		}
		scanIDs = append(scanIDs, scanID)
	}
	return scanIDs, rows.Err()
}

type IOTAReceipt struct {
	TxRef          string     `json:"txRef"`
	RecordID       string     `json:"recordId"`
	Signer         string     `json:"signer"`
	IncludedAt     *time.Time `json:"includedAt"`
	FinalizedAt    *time.Time `json:"finalizedAt"`
	DeleteUnlockAt *time.Time `json:"deleteUnlockAt"`
	Checkpoint     string     `json:"checkpoint"`
	GasComputation string     `json:"gasComputation"`
	GasStorage     string     `json:"gasStorage"`
	GasRebate      string     `json:"gasRebate"`
	TxSize         uint64     `json:"txSize"`
	RecordBytes    int        `json:"recordBytes"`
}

func (s *SQLiteIndex) SaveIOTAReceipt(ctx context.Context, record AttestationRecord, receipt IOTAReceipt) error {
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO attestations (
			network, app, version, scan_id, artifact_hash, sbom_hash, scan_hash, status,
			tx_ref, record_id, submitted_at, included_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, 'included', ?, ?, ?, ?)
		ON CONFLICT(network, scan_id) DO UPDATE SET
			status = 'included', tx_ref = excluded.tx_ref, record_id = excluded.record_id,
			included_at = excluded.included_at
		WHERE attestations.app = excluded.app AND attestations.version = excluded.version
			AND attestations.artifact_hash = excluded.artifact_hash
			AND attestations.sbom_hash = excluded.sbom_hash AND attestations.scan_hash = excluded.scan_hash
			AND (COALESCE(attestations.record_id, '') = '' OR attestations.record_id = excluded.record_id)
			AND NOT EXISTS (
				SELECT 1 FROM iota_transactions t WHERE t.network = excluded.network
				AND t.scan_id = excluded.scan_id AND t.digest != excluded.tx_ref
			)
	`, record.Network, record.App, record.Version, record.ScanID, record.ArtifactHash, record.SBOMHash,
		record.ScanHash, receipt.TxRef, receipt.RecordID, record.SubmittedAt, receipt.IncludedAt)
	if err != nil {
		return err
	}
	if rows, err := result.RowsAffected(); err != nil || rows == 0 {
		return errors.Join(err, errors.New("IOTA scan ID conflicts with the existing record"))
	}
	return s.UpdateIOTAReceipt(ctx, record.Network, record.ScanID, receipt)
}

func (s *SQLiteIndex) UpdateIOTAReceipt(ctx context.Context, network, scanID string, receipt IOTAReceipt) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO iota_receipts (network, scan_id, finalized_at, delete_unlock_at, checkpoint, gas_computation,
			gas_storage, gas_rebate, tx_size)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(network, scan_id) DO UPDATE SET
			finalized_at = COALESCE(excluded.finalized_at, iota_receipts.finalized_at),
			delete_unlock_at = COALESCE(excluded.delete_unlock_at, iota_receipts.delete_unlock_at),
			checkpoint = COALESCE(NULLIF(excluded.checkpoint, ''), iota_receipts.checkpoint),
			gas_computation = excluded.gas_computation, gas_storage = excluded.gas_storage,
			gas_rebate = excluded.gas_rebate, tx_size = MAX(excluded.tx_size, iota_receipts.tx_size)
	`, network, scanID, receipt.FinalizedAt, receipt.DeleteUnlockAt, receipt.Checkpoint, receipt.GasComputation,
		receipt.GasStorage, receipt.GasRebate, receipt.TxSize)
	return err
}

func (s *SQLiteIndex) IOTARecordID(ctx context.Context, network, app, version string) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `
		SELECT record_id FROM attestations WHERE network = ? AND app = ? AND version = ?
		AND status = 'included' AND COALESCE(record_id, '') != ''
		ORDER BY included_at DESC, id DESC LIMIT 1
	`, network, app, version).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return id, err
}
