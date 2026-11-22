package chain

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/brii26/eth-iota-attestation/oracle/internal/store"
	"github.com/brii26/eth-iota-attestation/smart-contract/registry"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var ErrTransactionReverted = errors.New("transaction reverted on chain")

func (a *EthereumAdapter) writeDurableRecord(ctx context.Context, record Record) (*Receipt, error) {
	prepared, err := a.index.EthereumTransaction(ctx, a.name, record.ScanID)
	if err != nil {
		return nil, err
	}
	if prepared == nil {
		prepared, err = a.prepareDurableRecord(ctx, record)
		if err != nil {
			return nil, err
		}
	}
	return a.resumeDurableRecord(ctx, *prepared)
}

func (a *EthereumAdapter) prepareDurableRecord(ctx context.Context, record Record) (*store.EthereumTransaction, error) {
	client := a.multiRPC.CurrentClient()
	nonce, release, err := a.nonceMgr.Acquire(ctx, client)
	if err != nil {
		return nil, err
	}
	saved := false
	defer func() { release(saved) }()
	auth, err := bind.NewKeyedTransactorWithChainID(a.privKey, a.chainID)
	if err != nil {
		return nil, err
	}
	auth.Context, auth.Nonce, auth.NoSend = ctx, new(big.Int).SetUint64(nonce), true
	contract, err := registry.NewAttestationRegistryTransactor(a.contractAddr, client)
	if err != nil {
		return nil, err
	}
	tx, err := contract.AddRecord(auth, record.App, record.Version, record.ScanID,
		record.ArtifactHash, record.SBOMHash, record.ScanHash)
	if err != nil {
		return nil, fmt.Errorf("prepare Ethereum transaction: %w", err)
	}
	raw, err := tx.MarshalBinary()
	if err != nil {
		return nil, err
	}
	prepared := store.EthereumTransaction{ScanID: record.ScanID, Hash: tx.Hash().Hex(),
		Raw: raw, Sender: a.oracleAddr.Hex(), Nonce: nonce}
	if err := a.index.SaveEthereumTransaction(ctx, a.name, prepared); err != nil {
		return nil, err
	}
	saved = true
	return &prepared, nil
}

func (a *EthereumAdapter) resumeDurableRecord(ctx context.Context, prepared store.EthereumTransaction) (*Receipt, error) {
	var tx types.Transaction
	if err := tx.UnmarshalBinary(prepared.Raw); err != nil {
		return nil, err
	}
	if tx.Hash().Hex() != prepared.Hash || tx.ChainId().Cmp(a.chainID) != 0 ||
		tx.To() == nil || *tx.To() != a.contractAddr {
		return nil, errors.New("stored Ethereum transaction does not match the configured network and registry")
	}
	wait, cancel := context.WithTimeout(ctx, a.txTimeout)
	defer cancel()
	var receipt *types.Receipt
	err := a.multiRPC.ExecuteWithRetry(wait, func(client *ethclient.Client) error {
		attempt, cancel := context.WithTimeout(wait, 4*time.Second)
		defer cancel()
		var err error
		receipt, err = client.TransactionReceipt(attempt, tx.Hash())
		if err == nil {
			return nil
		}
		if !errors.Is(err, ethereum.NotFound) {
			return err
		}
		if _, _, err := client.TransactionByHash(attempt, tx.Hash()); err == nil {
			return nil
		} else if !errors.Is(err, ethereum.NotFound) {
			return err
		}
		err = client.SendTransaction(attempt, &tx)
		if err != nil && strings.Contains(strings.ToLower(err.Error()), "already known") {
			return nil
		}
		return err
	})
	if err != nil {
		return nil, errors.Join(ErrTransactionPending, err)
	}
	for receipt == nil {
		select {
		case <-wait.Done():
			return nil, errors.Join(ErrTransactionPending, wait.Err())
		case <-time.After(time.Second):
		}
		attempt, stop := context.WithTimeout(wait, 4*time.Second)
		var err error
		receipt, err = a.multiRPC.CurrentClient().TransactionReceipt(attempt, tx.Hash())
		stop()
		if err != nil && !errors.Is(err, ethereum.NotFound) {
			_ = a.multiRPC.Failover(wait)
		}
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		err := a.index.UpdateStatus(ctx, a.name, prepared.ScanID, "failed", prepared.Hash, "", nil)
		if err != nil {
			return nil, errors.Join(ErrTransactionPending, err)
		}
		return nil, ErrTransactionReverted
	}
	var recordID string
	for _, entry := range receipt.Logs {
		if entry.Address == a.contractAddr {
			if event, err := a.filterer.ParseRecordAdded(*entry); err == nil && event.ScanId == prepared.ScanID {
				recordID = event.RecordId.String()
				break
			}
		}
	}
	if recordID == "" {
		return nil, errors.Join(ErrTransactionPending, errors.New("Ethereum receipt is missing the matching record event"))
	}
	included := time.Now().UTC()
	if err := a.index.UpdateStatus(ctx, a.name, prepared.ScanID, "included", prepared.Hash, recordID, &included); err != nil {
		return nil, errors.Join(ErrTransactionPending, err)
	}
	return &Receipt{
		TxRef: prepared.Hash, RecordID: recordID, IncludedAt: included,
		BlockNumber: receipt.BlockNumber.Uint64(), GasUsed: receipt.GasUsed, TxSize: uint64(tx.Size()),
	}, nil
}

func (a *EthereumAdapter) Reconcile(ctx context.Context) error {
	pending, err := a.index.PendingEthereumTransactions(ctx, a.name)
	if err != nil {
		return err
	}
	for _, prepared := range pending {
		if _, err := a.resumeDurableRecord(ctx, prepared); err != nil {
			log.Printf("Ethereum transaction %s: %v", prepared.Hash, err)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	return nil
}
