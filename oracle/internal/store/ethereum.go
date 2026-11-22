package store

import (
	"context"
	"database/sql"
	"errors"
)

type EthereumTransaction struct {
	ScanID string
	Hash   string
	Raw    []byte
	Sender string
	Nonce  uint64
}

func (s *SQLiteIndex) SaveEthereumTransaction(ctx context.Context, network string, tx EthereumTransaction) error {
	snapshot, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer snapshot.Rollback()
	_, err = snapshot.ExecContext(ctx, `
		INSERT INTO ethereum_transactions (network, scan_id, hash, raw, sender, nonce) VALUES (?, ?, ?, ?, ?, ?)
	`, network, tx.ScanID, tx.Hash, tx.Raw, tx.Sender, tx.Nonce)
	if err != nil {
		return err
	}
	result, err := snapshot.ExecContext(ctx, `
		UPDATE attestations SET tx_ref = ? WHERE network = ? AND scan_id = ? AND status = 'pending'
	`, tx.Hash, network, tx.ScanID)
	if err != nil {
		return err
	}
	if count, err := result.RowsAffected(); err != nil || count != 1 {
		return errors.Join(err, errors.New("Ethereum transaction has no pending attestation"))
	}
	return snapshot.Commit()
}

func (s *SQLiteIndex) EthereumTransaction(ctx context.Context, network, scanID string) (*EthereumTransaction, error) {
	transaction := EthereumTransaction{ScanID: scanID}
	err := s.db.QueryRowContext(ctx, `
		SELECT hash, raw, sender, nonce FROM ethereum_transactions WHERE network = ? AND scan_id = ?
	`, network, scanID).Scan(&transaction.Hash, &transaction.Raw, &transaction.Sender, &transaction.Nonce)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &transaction, err
}

func (s *SQLiteIndex) EthereumNonceFloor(ctx context.Context, network, sender string) (uint64, error) {
	var nonce uint64
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(nonce) + 1, 0) FROM ethereum_transactions WHERE network = ? AND sender = ?
	`, network, sender).Scan(&nonce)
	return nonce, err
}

func (s *SQLiteIndex) PendingEthereumTransactions(ctx context.Context, network string) ([]EthereumTransaction, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.scan_id, t.hash, t.raw, t.sender, t.nonce
		FROM ethereum_transactions t JOIN attestations a ON a.network = t.network AND a.scan_id = t.scan_id
		WHERE t.network = ? AND a.status = 'pending' ORDER BY t.sender, t.nonce
	`, network)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []EthereumTransaction
	for rows.Next() {
		var tx EthereumTransaction
		if err := rows.Scan(&tx.ScanID, &tx.Hash, &tx.Raw, &tx.Sender, &tx.Nonce); err != nil {
			return nil, err
		}
		results = append(results, tx)
	}
	return results, rows.Err()
}
