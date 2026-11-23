package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type BadgeJob struct {
	Recipient string
	RecordID  string
}

func (s *SQLiteIndex) NextBadge(ctx context.Context, network string) (*BadgeJob, error) {
	var job BadgeJob
	err := s.db.QueryRowContext(ctx, `
		SELECT recipient, record_id FROM pending_badges
		WHERE network = ? AND next_attempt <= ?
		ORDER BY next_attempt, recipient, record_id LIMIT 1`, network, time.Now().UTC()).Scan(&job.Recipient, &job.RecordID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &job, err
}

func (s *SQLiteIndex) FinishBadge(ctx context.Context, network string, job BadgeJob, issued bool) error {
	if issued {
		_, err := s.db.ExecContext(ctx, "DELETE FROM pending_badges WHERE network = ? AND recipient = ? AND record_id = ?", network, job.Recipient, job.RecordID)
		return err
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE pending_badges SET next_attempt = ?
		WHERE network = ? AND recipient = ? AND record_id = ?`, time.Now().UTC().Add(time.Minute), network, job.Recipient, job.RecordID)
	return err
}
