package verify

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/brii26/eth-iota-attestation/oracle/internal/chain"
	"github.com/brii26/eth-iota-attestation/oracle/internal/model"
)

// BlobStore defines the storage reader for SBOM and vulnerability scan files.
type BlobStore interface {
	GetSBOM(ctx context.Context, key string) ([]byte, error)
	GetScan(ctx context.Context, key string) ([]byte, error)
}

// ArtifactFetcher defines the package registry reader for downloading software tarballs.
type ArtifactFetcher interface {
	FetchArtifact(ctx context.Context, app, version string) ([]byte, error)
}

// EngineConfig holds dependencies required to perform full end-to-end attestation verification.
type EngineConfig struct {
	BlobStore       BlobStore
	ArtifactFetcher ArtifactFetcher
	MaxGrypeDBAge   time.Duration
}

// Engine coordinates retrieving on-chain records, fetching raw off-chain files, re-hashing,
// and validating signer authorization and scan freshness.
type Engine struct {
	blobStore       BlobStore
	artifactFetcher ArtifactFetcher
	maxGrypeDBAge   time.Duration
}

// NewEngine creates an initialized Engine.
func NewEngine(cfg EngineConfig) *Engine {
	if cfg.MaxGrypeDBAge == 0 {
		cfg.MaxGrypeDBAge = 48 * time.Hour
	}
	return &Engine{
		blobStore:       cfg.BlobStore,
		artifactFetcher: cfg.ArtifactFetcher,
		maxGrypeDBAge:   cfg.MaxGrypeDBAge,
	}
}

// Verify evaluates the integrity and authenticity of software artifacts for a given request.
func (e *Engine) Verify(ctx context.Context, adapter chain.Adapter, req model.VerifyRequest) (*model.VerifyResponse, error) {
	// 1. Resolve on-chain record
	var q chain.Query
	if req.RecordID != "" {
		var id *big.Int
		var ok bool
		if strings.HasPrefix(req.RecordID, "0x") || strings.HasPrefix(req.RecordID, "0X") {
			id, ok = new(big.Int).SetString(req.RecordID[2:], 16)
		} else {
			id, ok = new(big.Int).SetString(req.RecordID, 10)
			if !ok {
				id, ok = new(big.Int).SetString(req.RecordID, 16)
			}
		}
		if ok {
			q.RecordID = id
		}
	}
	q.App = req.App
	q.Version = req.Version

	chainRec, err := adapter.ReadRecord(ctx, q)
	if err != nil {
		if errors.Is(err, chain.ErrRecordNotFound) {
			return &model.VerifyResponse{
				Network:  req.Network,
				App:      req.App,
				Version:  req.Version,
				RecordID: req.RecordID,
				Status:   model.StatusRecordNotFound,
				Checks: model.VerifyChecks{
					Artifact:      model.CheckUnavailable,
					SBOM:          model.CheckUnavailable,
					Scan:          model.CheckUnavailable,
					SignerTrusted: false,
				},
			}, nil
		}
		return nil, fmt.Errorf("read on-chain record: %w", err)
	}

	checks := model.VerifyChecks{
		Artifact:      model.CheckMatch,
		SBOM:          model.CheckMatch,
		Scan:          model.CheckMatch,
		SignerTrusted: true,
	}

	var fileUnavailable bool
	var artifactMismatch bool
	var sbomMismatch bool
	var scanMismatch bool
	var isStale bool

	// 2. Fetch and verify artifact from Verdaccio
	if e.artifactFetcher != nil {
		artBytes, err := e.artifactFetcher.FetchArtifact(ctx, chainRec.App, chainRec.Version)
		if err != nil {
			checks.Artifact = model.CheckUnavailable
			fileUnavailable = true
		} else {
			calcArtHash, _ := ComputeSHA256(artBytes)
			if subtle.ConstantTimeCompare(calcArtHash[:], chainRec.ArtifactHash[:]) != 1 {
				checks.Artifact = model.CheckMismatch
				artifactMismatch = true
			}
		}
	} else {
		checks.Artifact = model.CheckUnavailable
		fileUnavailable = true
	}

	// 3. Fetch and verify SBOM & scan from RustFS
	storageKey := fmt.Sprintf("%s/%s/%s/%s.json", req.Network, chainRec.App, chainRec.Version, chainRec.ScanID)

	if e.blobStore != nil {
		sbomBytes, err := e.blobStore.GetSBOM(ctx, storageKey)
		if err != nil {
			checks.SBOM = model.CheckUnavailable
			fileUnavailable = true
		} else {
			calcSBOMHash, _ := ComputeSHA256(sbomBytes)
			if subtle.ConstantTimeCompare(calcSBOMHash[:], chainRec.SBOMHash[:]) != 1 {
				checks.SBOM = model.CheckMismatch
				sbomMismatch = true
			}
		}

		scanBytes, err := e.blobStore.GetScan(ctx, storageKey)
		if err != nil {
			checks.Scan = model.CheckUnavailable
			fileUnavailable = true
		} else {
			calcScanHash, _ := ComputeSHA256(scanBytes)
			if subtle.ConstantTimeCompare(calcScanHash[:], chainRec.ScanHash[:]) != 1 {
				checks.Scan = model.CheckMismatch
				scanMismatch = true
			}

			// Validate Grype scan freshness (does NOT alter integrity status, only sets stale flag)
			if _, err := ValidateGrypeFreshness(scanBytes, e.maxGrypeDBAge, time.Now().UTC()); err != nil {
				isStale = true
			}
		}
	} else {
		checks.SBOM = model.CheckUnavailable
		checks.Scan = model.CheckUnavailable
		fileUnavailable = true
	}

	// 4. Verify signer authorization on chain
	signerTrusted, err := adapter.IsSignerAuthorized(ctx, chainRec.Signer)
	if err != nil {
		// ponytail: assume untrusted if check fails
		signerTrusted = false
	}
	checks.SignerTrusted = signerTrusted

	// 5. Evaluate overall integrity status based on precedence
	var overallStatus string
	if fileUnavailable {
		overallStatus = model.StatusFileUnavailable
	} else if artifactMismatch {
		overallStatus = model.StatusArtifactMismatch
	} else if sbomMismatch {
		overallStatus = model.StatusSBOMMismatch
	} else if scanMismatch {
		overallStatus = model.StatusScanMismatch
	} else if !checks.SignerTrusted {
		overallStatus = model.StatusUntrustedSigner
	} else {
		overallStatus = model.StatusValid
	}

	recTime := chainRec.Timestamp

	return &model.VerifyResponse{
		Network:    req.Network,
		App:        chainRec.App,
		Version:    chainRec.Version,
		RecordID:   chainRec.RecordID,
		Status:     overallStatus,
		Stale:      isStale,
		Checks:     checks,
		Signer:     chainRec.Signer,
		RecordedAt: &recTime,
	}, nil
}
