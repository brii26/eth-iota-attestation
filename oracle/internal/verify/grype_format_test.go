package verify

import (
	"errors"
	"testing"
	"time"
)

func TestGrypeDatabaseStatus(t *testing.T) {
	now := time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)
	const recent = `{"descriptor":{"db":{"status":{"built":"2026-11-23T06:00:00Z","valid":true}}}}`
	const invalid = `{"descriptor":{"db":{"status":{"built":"2026-11-23T06:00:00Z","valid":false}}}}`
	const outdated = `{"descriptor":{"db":{"status":{"built":"2026-11-21T00:00:00Z","valid":true}}}}`
	built, err := ValidateGrypeFreshness([]byte(recent), 48*time.Hour, now)
	if err != nil || !built.Equal(now.Add(-6*time.Hour)) {
		t.Fatalf("current Grype format was rejected: %v, %v", built, err)
	}
	if _, err := ValidateGrypeFreshness([]byte(invalid), 48*time.Hour, now); err == nil {
		t.Fatal("invalid database was accepted")
	}
	if _, err := ValidateGrypeFreshness([]byte(outdated), 48*time.Hour, now); !errors.Is(err, ErrStaleScan) {
		t.Fatalf("want stale error, got %v", err)
	}
}
