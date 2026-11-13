package model

import (
	"encoding/json"
	"time"
)

// ClaimedHashes holds the hex-encoded SHA-256 hashes claimed by the pipeline.
type ClaimedHashes struct {
	Artifact string `json:"artifact"`
	SBOM     string `json:"sbom"`
	Scan     string `json:"scan"`
}

// AttestationRequest represents the incoming submission to POST /api/v1/attestations.
type AttestationRequest struct {
	Network       string          `json:"network"`
	App           string          `json:"app"`
	Version       string          `json:"version"`
	ScanID        string          `json:"scanId"`
	ClaimedHashes ClaimedHashes   `json:"claimedHashes"`
	SBOM          json.RawMessage `json:"sbom"`
	Scan          json.RawMessage `json:"scan"`
}

// AttestationStatus represents the response for GET /api/v1/attestations/{network}/{scanId}.
type AttestationStatus struct {
	Network     string     `json:"network"`
	ScanID      string     `json:"scanId"`
	Status      string     `json:"status"` // e.g. "pending", "included", "failed"
	TxRef       string     `json:"txRef,omitempty"`
	RecordID    string     `json:"recordId,omitempty"`
	SubmittedAt time.Time  `json:"submittedAt"`
	IncludedAt  *time.Time `json:"includedAt,omitempty"`
}

// ErrorDetail contains the structured code and message.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse formats all oracle API errors as { "error": { "code": "...", "message": "..." } }.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// Standard error codes specified by Issue #8.
const (
	ErrCodeInvalidInput          = "ERR_INVALID_INPUT"
	ErrCodeUnauthorizedPipeline  = "ERR_UNAUTHORIZED_PIPELINE"
	ErrCodeDuplicateScanID       = "ERR_DUPLICATE_SCAN_ID"
	ErrCodeStaleScan             = "ERR_STALE_SCAN"
	ErrCodeDependencyUnavailable = "ERR_DEPENDENCY_UNAVAILABLE"
	ErrCodeInternal              = "ERR_INTERNAL"
	ErrCodeNotFound              = "ERR_NOT_FOUND"
)
