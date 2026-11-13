package verify

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// grypeDescriptor represents the relevant metadata block in a Grype JSON scan report.
type grypeReport struct {
	Descriptor struct {
		DB struct {
			Built string `json:"built"`
		} `json:"db"`
	} `json:"descriptor"`
}

// ErrMissingDBTimestamp is returned when the scan payload lacks the descriptor.db.built field.
var ErrMissingDBTimestamp = errors.New("missing descriptor.db.built in Grype scan")

// ErrStaleScan is returned when the Grype DB build timestamp is older than maxAge.
var ErrStaleScan = errors.New("vulnerability database is stale")

// ValidateGrypeFreshness parses the descriptor.db.built field and ensures it was built within maxAge.
func ValidateGrypeFreshness(scanBytes []byte, maxAge time.Duration, now time.Time) (time.Time, error) {
	var report grypeReport
	if err := json.Unmarshal(scanBytes, &report); err != nil {
		return time.Time{}, fmt.Errorf("unmarshal Grype scan: %w", err)
	}

	builtStr := report.Descriptor.DB.Built
	if builtStr == "" {
		return time.Time{}, ErrMissingDBTimestamp
	}

	// Try RFC3339 / ISO 8601
	builtTime, err := time.Parse(time.RFC3339, builtStr)
	if err != nil {
		// Fallback for timestamps without timezone offset
		builtTime, err = time.Parse("2006-01-02T15:04:05", builtStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("parse descriptor.db.built timestamp %q: %w", builtStr, err)
		}
	}

	if maxAge > 0 && now.Sub(builtTime) > maxAge {
		return builtTime, fmt.Errorf("%w: built at %s, exceeds max age %s", ErrStaleScan, builtTime.Format(time.RFC3339), maxAge)
	}

	return builtTime, nil
}
