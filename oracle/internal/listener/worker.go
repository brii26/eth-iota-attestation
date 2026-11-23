package listener

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/brii26/eth-iota-attestation/oracle/internal/chain"
	"github.com/brii26/eth-iota-attestation/oracle/internal/model"
	"github.com/brii26/eth-iota-attestation/oracle/internal/store"
	"github.com/brii26/eth-iota-attestation/oracle/internal/verify"
)

const integrityValid uint8 = 1

// StatusToIntegrityStatus maps domain status string to contract IntegrityStatus enum value.
// Solidity IntegrityStatus:
// 0: Pending, 1: Valid, 2: ArtifactMismatch, 3: SbomMismatch, 4: ScanMismatch,
// 5: UntrustedSigner, 6: RecordNotFound, 7: FileUnavailable.
func StatusToIntegrityStatus(statusStr string, isStale bool) (uint8, bool) {
	switch statusStr {
	case model.StatusValid:
		return integrityValid, isStale
	case model.StatusArtifactMismatch:
		return 2, false
	case model.StatusSBOMMismatch:
		return 3, false
	case model.StatusScanMismatch:
		return 4, false
	case model.StatusUntrustedSigner:
		return 5, false
	case model.StatusRecordNotFound:
		return 6, false
	case model.StatusFileUnavailable:
		return 7, false
	default:
		return 6, false
	}
}

// Config specifies runtime parameters for the on-chain verification listener worker.
type Config struct {
	Adapter      chain.Adapter
	Engine       *verify.Engine
	Index        *store.SQLiteIndex
	Network      string
	PollInterval time.Duration
	StartBlock   uint64
	MaxRetries   int
	RetryBackoff time.Duration
	MintBadge    chain.BadgeMinter
}

// Worker continuously polls for VerificationRequested events and fulfills them on-chain.
type Worker struct {
	cfg Config
}

// NewWorker initializes a new verification listener worker.
func NewWorker(cfg Config) *Worker {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 2 * time.Second
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	if cfg.RetryBackoff <= 0 {
		cfg.RetryBackoff = 500 * time.Millisecond
	}
	return &Worker{
		cfg: cfg,
	}
}

// PollOnce checks for new VerificationRequested events since last processed block,
// verifies their records, and fulfills them on-chain.
func (w *Worker) PollOnce(ctx context.Context) (int, error) {
	if w.cfg.MintBadge != nil {
		defer w.issueBadge(ctx)
	}
	// 1. Recover last processed block
	lastBlock, err := w.cfg.Index.GetLastProcessedBlock(ctx, w.cfg.Network)
	if err != nil {
		return 0, fmt.Errorf("get last processed block: %w", err)
	}

	if lastBlock == 0 && w.cfg.StartBlock > 0 {
		lastBlock = w.cfg.StartBlock - 1
	}

	currentBlock, err := w.cfg.Adapter.CurrentBlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("get current block: %w", err)
	}

	if currentBlock <= lastBlock {
		return 0, nil
	}

	start := lastBlock + 1
	end := currentBlock

	// 2. Query events
	events, err := w.cfg.Adapter.FilterVerificationRequests(ctx, &start, &end)
	if err != nil {
		return 0, fmt.Errorf("filter verification requests [%d, %d]: %w", start, end, err)
	}

	processedCount := 0
	for _, ev := range events {
		// 3. Replay-protection / idempotency check
		reqIDStr := ev.RequestID.String()
		alreadyDone, err := w.cfg.Index.IsVerificationFulfilled(ctx, w.cfg.Network, reqIDStr)
		if err != nil {
			return processedCount, fmt.Errorf("check fulfilled status: %w", err)
		}
		if alreadyDone {
			continue
		}

		// 4. Run verification logic
		recordIDStr := ""
		if ev.RecordID != nil {
			recordIDStr = ev.RecordID.String()
		}

		result, txRef, err := w.fulfillEvent(ctx, ev.RequestID, recordIDStr)
		if err != nil {
			return processedCount, err
		}

		recipient := ""
		if w.cfg.MintBadge != nil && result.Status == integrityValid {
			recipient = ev.Requester
		}
		if err := w.cfg.Index.RecordFulfilledVerification(ctx, w.cfg.Network, reqIDStr, recordIDStr, result.Status, result.Stale, txRef, recipient); err != nil {
			return processedCount, fmt.Errorf("record fulfilled verification: %w", err)
		}

		processedCount++
	}

	// 7. Update checkpoint block
	if err := w.cfg.Index.SetLastProcessedBlock(ctx, w.cfg.Network, end); err != nil {
		return processedCount, fmt.Errorf("set last processed block: %w", err)
	}

	return processedCount, nil
}

func (w *Worker) fulfillEvent(ctx context.Context, requestID *big.Int, recordID string) (chain.VerificationResult, string, error) {
	reader, recoverable := w.cfg.Adapter.(interface {
		ReadVerificationResult(context.Context, *big.Int) (chain.VerificationResult, bool, error)
	})
	if recoverable {
		result, fulfilled, err := reader.ReadVerificationResult(ctx, requestID)
		if fulfilled || err != nil {
			return result, "", err
		}
	}
	response, err := w.cfg.Engine.Verify(ctx, w.cfg.Adapter, model.VerifyRequest{Network: w.cfg.Network, RecordID: recordID})
	if err != nil {
		return chain.VerificationResult{}, "", fmt.Errorf("verify record %s: %w", recordID, err)
	}
	status, stale := StatusToIntegrityStatus(response.Status, response.Stale)
	receipt, err := w.fulfillWithRetry(ctx, requestID, status, stale)
	if err != nil {
		return chain.VerificationResult{}, "", fmt.Errorf("fulfill verification request %s: %w", requestID, err)
	}
	txRef := ""
	if receipt != nil {
		txRef = receipt.TxRef
	}
	return chain.VerificationResult{Status: status, Stale: stale}, txRef, nil
}

func (w *Worker) issueBadge(ctx context.Context) {
	job, err := w.cfg.Index.NextBadge(ctx, w.cfg.Network)
	if err != nil {
		log.Printf("Read pending badge on %s: %v", w.cfg.Network, err)
		return
	}
	if job == nil {
		return
	}
	recordID, ok := new(big.Int).SetString(job.RecordID, 10)
	if !ok {
		log.Printf("Invalid record ID for pending badge on %s", w.cfg.Network)
	} else {
		err = w.cfg.MintBadge(ctx, job.Recipient, recordID)
		if err != nil {
			log.Printf("Mint badge on %s for record %s: %v", w.cfg.Network, job.RecordID, err)
		}
	}
	if finishErr := w.cfg.Index.FinishBadge(ctx, w.cfg.Network, *job, ok && err == nil); finishErr != nil {
		log.Printf("Update pending badge on %s: %v", w.cfg.Network, finishErr)
	}
}

func (w *Worker) fulfillWithRetry(ctx context.Context, requestID *big.Int, status uint8, stale bool) (*chain.Receipt, error) {
	var lastErr error
	backoff := w.cfg.RetryBackoff

	for attempt := 0; attempt < w.cfg.MaxRetries; attempt++ {
		receipt, err := w.cfg.Adapter.FulfillVerification(ctx, requestID, status, stale)
		if err == nil {
			return receipt, nil
		}
		lastErr = err

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
		}
	}
	return nil, fmt.Errorf("fulfillment failed after %d attempts: %w", w.cfg.MaxRetries, lastErr)
}

// Start runs the worker event loop in the foreground until the context is cancelled.
func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.cfg.PollInterval)
	defer ticker.Stop()

	// Initial poll on startup to catch up missed requests immediately
	_, _ = w.PollOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = w.PollOnce(ctx)
		}
	}
}
