package model

import (
	"time"
)

// Integrity statuses specified by Issue #10.
const (
	StatusValid             = "VALID"
	StatusArtifactMismatch  = "ARTIFACT_MISMATCH"
	StatusSBOMMismatch      = "SBOM_MISMATCH"
	StatusScanMismatch      = "SCAN_MISMATCH"
	StatusUntrustedSigner   = "UNTRUSTED_SIGNER"
	StatusRecordNotFound    = "RECORD_NOT_FOUND"
	StatusFileUnavailable   = "FILE_UNAVAILABLE"

	CheckMatch       = "MATCH"
	CheckMismatch    = "MISMATCH"
	CheckUnavailable = "UNAVAILABLE"
)

// VerifyRequest represents the request body for POST /api/v1/verify.
type VerifyRequest struct {
	Network  string `json:"network"`
	App      string `json:"app"`
	Version  string `json:"version"`
	RecordID string `json:"recordId,omitempty"` // optional; verifies latest if empty
}

// VerifyChecks provides granular breakdown of artifact, SBOM, scan, and signer verification.
type VerifyChecks struct {
	Artifact      string `json:"artifact"`
	SBOM          string `json:"sbom"`
	Scan          string `json:"scan"`
	SignerTrusted bool   `json:"signerTrusted"`
}

// VerifyResponse represents the structured result of POST /api/v1/verify.
type VerifyResponse struct {
	Network    string       `json:"network"`
	App        string       `json:"app"`
	Version    string       `json:"version"`
	RecordID   string       `json:"recordId,omitempty"`
	Status     string       `json:"status"`
	Stale      bool         `json:"stale"`
	Checks     VerifyChecks `json:"checks"`
	Signer     string       `json:"signer,omitempty"`
	RecordedAt *time.Time   `json:"recordedAt,omitempty"`
}

// RecordItem represents an entry in the record history list.
type RecordItem struct {
	RecordID     string     `json:"recordId"`
	Network      string     `json:"network"`
	App          string     `json:"app"`
	Version      string     `json:"version"`
	ScanID       string     `json:"scanId"`
	ArtifactHash string     `json:"artifactHash"`
	SBOMHash     string     `json:"sbomHash"`
	ScanHash     string     `json:"scanHash"`
	Status       string     `json:"status"`
	TxRef        string     `json:"txRef"`
	SubmittedAt  time.Time  `json:"submittedAt"`
	IncludedAt   *time.Time `json:"includedAt,omitempty"`
}

// RecordHistoryResponse represents the response for GET /api/v1/records.
type RecordHistoryResponse struct {
	Network string       `json:"network"`
	App     string       `json:"app"`
	Version string       `json:"version"`
	Total   int          `json:"total"`
	Records []RecordItem `json:"records"`
}
