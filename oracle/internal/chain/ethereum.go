package chain

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/brii26/eth-iota-attestation/smart-contract/registry"
)

// EthereumConfig defines options required to connect to an EVM network.
type EthereumConfig struct {
	Name            string
	Endpoints       []string // Primary and backup RPC endpoints
	ContractAddress string
	PrivateKeyHex   string
	TxTimeout       time.Duration
}

// EthereumAdapter implements the Adapter interface for EVM networks (private Geth, Sepolia).
type EthereumAdapter struct {
	name         string
	multiRPC     *MultiRPCClient
	contractAddr common.Address
	privKey      *ecdsa.PrivateKey
	oracleAddr   common.Address
	chainID      *big.Int
	nonceMgr     *NonceManager
	transactor   *registry.AttestationRegistryTransactor
	caller       *registry.AttestationRegistryCaller
	filterer     *registry.AttestationRegistryFilterer
	txTimeout    time.Duration
}

// NewEthereumAdapter initializes Ethereum RPC connections, contract bindings, and the nonce coordinator.
func NewEthereumAdapter(ctx context.Context, cfg EthereumConfig) (*EthereumAdapter, error) {
	if cfg.Name == "" {
		cfg.Name = "ethereum-private"
	}
	if len(cfg.Endpoints) == 0 {
		return nil, errors.New("at least one Ethereum RPC endpoint is required")
	}
	if cfg.ContractAddress == "" {
		return nil, errors.New("contract address is required")
	}
	if cfg.TxTimeout == 0 {
		cfg.TxTimeout = 60 * time.Second
	}

	contractAddr := common.HexToAddress(cfg.ContractAddress)

	// Clean and parse private key
	cleanKey := strings.TrimPrefix(strings.TrimSpace(cfg.PrivateKeyHex), "0x")
	key, err := crypto.HexToECDSA(cleanKey)
	if err != nil {
		return nil, fmt.Errorf("invalid oracle private key: %w", err)
	}
	oracleAddr := crypto.PubkeyToAddress(key.PublicKey)

	multiRPC, err := NewMultiRPCClient(ctx, cfg.Endpoints)
	if err != nil {
		return nil, fmt.Errorf("init multi-rpc client: %w", err)
	}

	client := multiRPC.CurrentClient()
	chainID, err := client.ChainID(ctx)
	if err != nil {
		multiRPC.Close()
		return nil, fmt.Errorf("fetch chain ID: %w", err)
	}

	transactor, err := registry.NewAttestationRegistryTransactor(contractAddr, client)
	if err != nil {
		multiRPC.Close()
		return nil, fmt.Errorf("bind transactor: %w", err)
	}

	caller, err := registry.NewAttestationRegistryCaller(contractAddr, client)
	if err != nil {
		multiRPC.Close()
		return nil, fmt.Errorf("bind caller: %w", err)
	}

	filterer, err := registry.NewAttestationRegistryFilterer(contractAddr, client)
	if err != nil {
		multiRPC.Close()
		return nil, fmt.Errorf("bind filterer: %w", err)
	}

	return &EthereumAdapter{
		name:         cfg.Name,
		multiRPC:     multiRPC,
		contractAddr: contractAddr,
		privKey:      key,
		oracleAddr:   oracleAddr,
		chainID:      chainID,
		nonceMgr:     NewNonceManager(oracleAddr),
		transactor:   transactor,
		caller:       caller,
		filterer:     filterer,
		txTimeout:    cfg.TxTimeout,
	}, nil
}

// Name returns the identifier of this adapter.
func (a *EthereumAdapter) Name() string {
	return a.name
}

// ContractAddress returns the Ethereum address of the registry contract.
func (a *EthereumAdapter) ContractAddress() common.Address {
	return a.contractAddr
}

// Filterer returns the contract event filterer.
func (a *EthereumAdapter) Filterer() *registry.AttestationRegistryFilterer {
	return a.filterer
}

// WriteRecord writes a new attestation on-chain by calling addRecord on AttestationRegistry.
func (a *EthereumAdapter) WriteRecord(ctx context.Context, rec Record) (*Receipt, error) {
	client := a.multiRPC.CurrentClient()

	nonce, releaseNonce, err := a.nonceMgr.Acquire(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("acquire nonce: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(a.privKey, a.chainID)
	if err != nil {
		releaseNonce(false)
		return nil, fmt.Errorf("create transactor: %w", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Context = ctx

	// Query suggested gas tip cap and fee
	tipCap, err := client.SuggestGasTipCap(ctx)
	if err == nil {
		auth.GasTipCap = tipCap
	}

	tx, err := a.transactor.AddRecord(
		auth,
		rec.App,
		rec.Version,
		rec.ScanID,
		rec.ArtifactHash,
		rec.SBOMHash,
		rec.ScanHash,
	)
	if err != nil {
		releaseNonce(false)
		return nil, fmt.Errorf("send addRecord transaction: %w", err)
	}

	// Transaction successfully broadcast into the mempool
	releaseNonce(true)
	submittedAt := time.Now().UTC()

	// Wait for inclusion and receipt
	waitCtx, cancel := context.WithTimeout(ctx, a.txTimeout)
	defer cancel()

	rawReceipt, err := bind.WaitMined(waitCtx, client, tx)
	if err != nil {
		return nil, fmt.Errorf("wait for transaction %s: %w", tx.Hash().Hex(), err)
	}

	if rawReceipt.Status != types.ReceiptStatusSuccessful {
		return nil, fmt.Errorf("transaction %s reverted on chain", tx.Hash().Hex())
	}

	// Extract recordId from the RecordAdded event log
	var recordIDStr string
	for _, l := range rawReceipt.Logs {
		if l.Address == a.contractAddr {
			event, err := a.filterer.ParseRecordAdded(*l)
			if err == nil && event != nil {
				recordIDStr = event.RecordId.String()
				break
			}
		}
	}
	includedAt := time.Now().UTC()

	return &Receipt{
		TxRef:       tx.Hash().Hex(),
		RecordID:    recordIDStr,
		BlockNumber: rawReceipt.BlockNumber.Uint64(),
		GasUsed:     rawReceipt.GasUsed,
		TxSize:      uint64(tx.Size()),
		SubmittedAt: submittedAt,
		IncludedAt:  includedAt,
	}, nil
}

// ReadRecord looks up the latest attestation record for the given app and version.
func (a *EthereumAdapter) ReadRecord(ctx context.Context, query Query) (*ChainRecord, error) {
	opts := &bind.CallOpts{Context: ctx}

	var rec registry.AttestationRegistryRecord
	var err error
	var recordIDStr string

	if query.RecordID != nil {
		rec, err = a.caller.GetRecord(opts, query.RecordID)
		recordIDStr = query.RecordID.String()
	} else if query.App != "" && query.Version != "" {
		rec, err = a.caller.GetLatestRecord(opts, query.App, query.Version)
		if err == nil {
			ids, idErr := a.caller.GetRecordIds(opts, query.App, query.Version, big.NewInt(0), big.NewInt(100))
			if idErr == nil && len(ids) > 0 {
				recordIDStr = ids[len(ids)-1].String()
			}
		}
	} else {
		return nil, errors.New("query must specify either recordId or both app and version")
	}

	if err != nil {
		return nil, fmt.Errorf("call registry: %w", err)
	}

	if rec.Timestamp == nil || rec.Timestamp.Sign() == 0 {
		return nil, ErrRecordNotFound
	}

	return &ChainRecord{
		RecordID:     recordIDStr,
		App:          rec.App,
		Version:      rec.Version,
		ScanID:       rec.ScanId,
		ArtifactHash: rec.ArtifactHash,
		SBOMHash:     rec.SbomHash,
		ScanHash:     rec.ScanHash,
		Signer:       rec.Signer.Hex(),
		Timestamp:    time.Unix(rec.Timestamp.Int64(), 0).UTC(),
	}, nil
}

// IsSignerAuthorized checks if the signer is currently an authorized oracle on-chain.
func (a *EthereumAdapter) IsSignerAuthorized(ctx context.Context, signer string) (bool, error) {
	addr := common.HexToAddress(signer)
	return a.caller.IsAuthorizedOracle(&bind.CallOpts{Context: ctx}, addr)
}

// CurrentBlockNumber queries the latest block number from the primary RPC client.
func (a *EthereumAdapter) CurrentBlockNumber(ctx context.Context) (uint64, error) {
	client := a.multiRPC.CurrentClient()
	return client.BlockNumber(ctx)
}

// FilterVerificationRequests queries VerificationRequested events between startBlock and endBlock.
func (a *EthereumAdapter) FilterVerificationRequests(ctx context.Context, startBlock, endBlock *uint64) ([]VerificationEvent, error) {
	opts := &bind.FilterOpts{
		Context: ctx,
	}
	if startBlock != nil {
		opts.Start = *startBlock
	}
	opts.End = endBlock

	iter, err := a.filterer.FilterVerificationRequested(opts, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("filter verification requested events: %w", err)
	}
	defer iter.Close()

	var events []VerificationEvent
	for iter.Next() {
		ev := iter.Event
		if ev == nil {
			continue
		}
		events = append(events, VerificationEvent{
			RequestID:   ev.RequestId,
			RecordID:    ev.RecordId,
			Requester:   ev.Requester.Hex(),
			BlockNumber: ev.Raw.BlockNumber,
		})
	}
	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("iterator error: %w", err)
	}
	return events, nil
}

// FulfillVerification submits a verification result for a request on-chain.
func (a *EthereumAdapter) FulfillVerification(ctx context.Context, requestID *big.Int, status uint8, stale bool) (*Receipt, error) {
	client := a.multiRPC.CurrentClient()

	nonce, releaseNonce, err := a.nonceMgr.Acquire(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("acquire nonce: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(a.privKey, a.chainID)
	if err != nil {
		releaseNonce(false)
		return nil, fmt.Errorf("create transactor: %w", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Context = ctx

	tipCap, err := client.SuggestGasTipCap(ctx)
	if err == nil {
		auth.GasTipCap = tipCap
	}

	result := registry.AttestationRegistryVerificationResult{
		Status: status,
		Stale:  stale,
	}

	tx, err := a.transactor.FulfillVerification(auth, requestID, result)
	if err != nil {
		releaseNonce(false)
		return nil, fmt.Errorf("send fulfillVerification transaction: %w", err)
	}

	releaseNonce(true)
	submittedAt := time.Now().UTC()

	waitCtx, cancel := context.WithTimeout(ctx, a.txTimeout)
	defer cancel()

	rawReceipt, err := bind.WaitMined(waitCtx, client, tx)
	if err != nil {
		return nil, fmt.Errorf("wait for fulfill transaction %s: %w", tx.Hash().Hex(), err)
	}

	if rawReceipt.Status != types.ReceiptStatusSuccessful {
		return nil, fmt.Errorf("fulfill transaction %s reverted on chain", tx.Hash().Hex())
	}

	includedAt := time.Now().UTC()
	return &Receipt{
		TxRef:       tx.Hash().Hex(),
		RecordID:    requestID.String(),
		BlockNumber: rawReceipt.BlockNumber.Uint64(),
		GasUsed:     rawReceipt.GasUsed,
		TxSize:      uint64(tx.Size()),
		SubmittedAt: submittedAt,
		IncludedAt:  includedAt,
	}, nil
}

// Close releases the RPC connections.
func (a *EthereumAdapter) Close() {
	a.multiRPC.Close()
}
