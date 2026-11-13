package chain

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/brii26/eth-iota-attestation/oracle/internal/store"
)

// RebuildIndexFromEvents scans on-chain RecordAdded events and populates missing records
// into the SQLite index, ensuring the index can always be fully reconstructed from the blockchain.
func RebuildIndexFromEvents(ctx context.Context, adapter *EthereumAdapter, idx *store.SQLiteIndex, startBlock uint64) (int, error) {
	opts := &bind.FilterOpts{
		Start:   startBlock,
		Context: ctx,
	}

	it, err := adapter.Filterer().FilterRecordAdded(opts, nil, nil, nil)
	if err != nil {
		return 0, fmt.Errorf("filter RecordAdded events: %w", err)
	}
	defer it.Close()

	rebuiltCount := 0
	callOpts := &bind.CallOpts{Context: ctx}

	for it.Next() {
		ev := it.Event
		exists, err := idx.Exists(ctx, adapter.Name(), ev.ScanId)
		if err != nil {
			return rebuiltCount, fmt.Errorf("check event existence: %w", err)
		}
		if exists {
			continue
		}

		// Read full record to get hashes
		onChainRec, err := adapter.caller.GetRecord(callOpts, ev.RecordId)
		if err != nil {
			return rebuiltCount, fmt.Errorf("fetch record %s: %w", ev.RecordId, err)
		}

		timestamp := time.Unix(onChainRec.Timestamp.Int64(), 0).UTC()
		rec := store.AttestationRecord{
			Network:      adapter.Name(),
			App:          ev.App,
			Version:      ev.Version,
			ScanID:       ev.ScanId,
			ArtifactHash: hex.EncodeToString(onChainRec.ArtifactHash[:]),
			SBOMHash:     hex.EncodeToString(onChainRec.SbomHash[:]),
			ScanHash:     hex.EncodeToString(onChainRec.ScanHash[:]),
			Status:       "included",
			TxRef:        ev.Raw.TxHash.Hex(),
			RecordID:     ev.RecordId.String(),
			SubmittedAt:  timestamp,
			IncludedAt:   &timestamp,
		}

		if err := idx.Save(ctx, rec); err != nil {
			return rebuiltCount, fmt.Errorf("save rebuilt record %s: %w", ev.ScanId, err)
		}
		rebuiltCount++
	}

	if err := it.Error(); err != nil {
		return rebuiltCount, fmt.Errorf("iterate RecordAdded events: %w", err)
	}

	return rebuiltCount, nil
}
