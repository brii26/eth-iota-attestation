package listener

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/brii26/eth-iota-attestation/oracle/internal/chain"
	"github.com/brii26/eth-iota-attestation/oracle/internal/model"
	"github.com/brii26/eth-iota-attestation/oracle/internal/store"
	"github.com/brii26/eth-iota-attestation/oracle/internal/verify"
)

// StatusToIntegrityStatus maps domain status string to contract IntegrityStatus enum value.
// Solidity IntegrityStatus:
// 0: Pending, 1: Valid, 2: ArtifactMismatch, 3: SbomMismatch, 4: ScanMismatch,
// 5: UntrustedSigner, 6: RecordNotFound, 7: FileUnavailable.
func StatusToIntegrityStatus(statusStr string, isStale bool) (uint8, bool) {
	switch statusStr {
	case model.StatusValid:
		return 1, isStale
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

		resp, err := w.cfg.Engine.Verify(ctx, w.cfg.Adapter, model.VerifyRequest{
			Network:  w.cfg.Network,
			RecordID: recordIDStr,
		})
		if err != nil {
			return processedCount, fmt.Errorf("verify record %s: %w", recordIDStr, err)
		}

		statusUint, staleBool := StatusToIntegrityStatus(resp.Status, resp.Stale)

		// 5. Fulfill on-chain with retry
		receipt, err := w.fulfillWithRetry(ctx, ev.RequestID, statusUint, staleBool)
		if err != nil {
			return processedCount, fmt.Errorf("fulfill verification request %s: %w", reqIDStr, err)
		}

		// 6. Record fulfillment in SQLite
		txRef := ""
		if receipt != nil {
			txRef = receipt.TxRef
		}
		if err := w.cfg.Index.RecordFulfilledVerification(ctx, w.cfg.Network, reqIDStr, recordIDStr, statusUint, staleBool, txRef); err != nil {
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
