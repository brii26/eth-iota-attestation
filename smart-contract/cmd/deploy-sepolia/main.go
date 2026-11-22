package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/brii26/eth-iota-attestation/smart-contract/registry"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Minute)
	defer cancel()
	if err := deploy(ctx); err != nil {
		log.Fatal(err)
	}
}

func deploy(ctx context.Context) error {
	rpcURL := os.Getenv("ETHEREUM_SEPOLIA_RPC_URL")
	if rpcURL == "" {
		return errors.New("ETHEREUM_SEPOLIA_RPC_URL is required")
	}
	adminKey, err := crypto.HexToECDSA(strings.TrimPrefix(os.Getenv("ETHEREUM_ADMIN_PRIVATE_KEY"), "0x"))
	if err != nil {
		return errors.New("ETHEREUM_ADMIN_PRIVATE_KEY must contain a Sepolia signing key")
	}
	oracleKey, err := crypto.HexToECDSA(strings.TrimPrefix(os.Getenv("ETHEREUM_ORACLE_PRIVATE_KEY"), "0x"))
	if err != nil {
		return errors.New("ETHEREUM_ORACLE_PRIVATE_KEY must contain a Sepolia signing key")
	}
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return fmt.Errorf("connect Sepolia RPC: %w", err)
	}
	defer client.Close()
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return err
	}
	if chainID.Cmp(big.NewInt(11155111)) != 0 {
		return fmt.Errorf("unexpected chain ID %s, expected Sepolia 11155111", chainID)
	}
	admin, err := bind.NewKeyedTransactorWithChainID(adminKey, chainID)
	if err != nil {
		return err
	}
	oracle, err := bind.NewKeyedTransactorWithChainID(oracleKey, chainID)
	if err != nil {
		return err
	}
	if admin.From == oracle.From {
		return errors.New("use separate Sepolia admin and oracle accounts")
	}
	admin.Context, oracle.Context = ctx, ctx
	for _, address := range []common.Address{admin.From, oracle.From} {
		balance, err := client.BalanceAt(ctx, address, nil)
		if err != nil {
			return err
		}
		if balance.Sign() == 0 {
			return fmt.Errorf("Sepolia account %s has no faucet balance", address)
		}
	}

	address, deployment, contract, err := registry.DeployAttestationRegistry(admin, client)
	if err != nil {
		return err
	}
	log.Printf("Sepolia deployment %s, transaction %s", address.Hex(), deployment.Hash().Hex())
	if _, err := waitSuccessful(ctx, client, deployment); err != nil {
		return err
	}
	permission, err := contract.AuthorizeOracle(admin, oracle.From)
	if err != nil {
		return err
	}
	if _, err := waitSuccessful(ctx, client, permission); err != nil {
		return err
	}
	scanID := "sepolia-smoke-" + deployment.Hash().Hex()
	artifact := [32]byte{1}
	sbom := [32]byte{2}
	scan := [32]byte{3}
	smoke, err := contract.AddRecord(oracle, "sepolia-smoke", "1.0.0", scanID, artifact, sbom, scan)
	if err != nil {
		return err
	}
	log.Printf("Sepolia smoke transaction %s", smoke.Hash().Hex())
	receipt, err := waitSuccessful(ctx, client, smoke)
	if err != nil {
		return err
	}
	if err := waitFinalized(ctx, client, receipt); err != nil {
		return err
	}
	record, err := contract.GetLatestRecord(&bind.CallOpts{Context: ctx}, "sepolia-smoke", "1.0.0")
	if err != nil {
		return err
	}
	if record.ScanId != scanID || record.Signer != oracle.From ||
		record.ArtifactHash != artifact || record.SbomHash != sbom || record.ScanHash != scan {
		return errors.New("finalized record did not match the smoke submission")
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{
		"network": "ethereum-sepolia", "chainId": chainID.String(), "contractAddress": address,
		"administrator": admin.From, "oracle": oracle.From, "deploymentTx": deployment.Hash(),
		"permissionTx": permission.Hash(), "recordTx": smoke.Hash(), "recordBlock": receipt.BlockNumber,
		"gasUsed": receipt.GasUsed, "scanId": scanID, "status": "finalized",
	})
}

func waitSuccessful(ctx context.Context, client *ethclient.Client, tx *types.Transaction) (*types.Receipt, error) {
	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		return nil, fmt.Errorf("wait for transaction %s: %w", tx.Hash(), err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, fmt.Errorf("transaction %s reverted", tx.Hash())
	}
	return receipt, nil
}

func waitFinalized(ctx context.Context, client *ethclient.Client, included *types.Receipt) error {
	ticker := time.NewTicker(12 * time.Second)
	defer ticker.Stop()
	for {
		finalized, err := client.HeaderByNumber(ctx, big.NewInt(int64(rpc.FinalizedBlockNumber)))
		if err != nil && !errors.Is(err, ethereum.NotFound) {
			return fmt.Errorf("read finalized Sepolia block: %w", err)
		}
		if finalized != nil && finalized.Number.Cmp(included.BlockNumber) >= 0 {
			current, err := client.TransactionReceipt(ctx, included.TxHash)
			if err != nil {
				return fmt.Errorf("confirm finalized receipt: %w", err)
			}
			if current.BlockHash != included.BlockHash {
				return errors.New("transaction moved during a reorg")
			}
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for Sepolia finality: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}
