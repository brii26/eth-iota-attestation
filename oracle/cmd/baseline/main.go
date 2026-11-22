package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/brii26/eth-iota-attestation/oracle/internal/chain"
	"github.com/brii26/eth-iota-attestation/oracle/internal/store"
	"github.com/brii26/eth-iota-attestation/smart-contract/registry"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	network := flag.String("network", "ethereum-private", "Ethereum private, Sepolia, IOTA testnet or IOTA localnet")
	action := flag.String("action", "verify", "deploy, write or verify")
	reference := flag.String("record", "", "IOTA baseline notarization object ID")
	artifact := flag.String("artifact", "", "artifact SHA-256")
	sbom := flag.String("sbom", "", "SBOM SHA-256")
	scan := flag.String("scan", "", "scan SHA-256")
	flag.Parse()
	if *action != "deploy" && *action != "write" && *action != "verify" {
		log.Fatal("choose -action deploy, write or verify")
	}
	var hashes [3][32]byte
	if *action != "deploy" {
		for index, input := range []string{*artifact, *sbom, *scan} {
			data, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
			if err != nil || len(data) != 32 {
				log.Fatal("all three file fingerprints must be 32-byte hex strings")
			}
			copy(hashes[index][:], data)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	var err error
	switch *network {
	case "ethereum-private":
		err = ethereum(ctx, 1337, *action, hashes)
	case "ethereum-sepolia":
		err = ethereum(ctx, 11155111, *action, hashes)
	case "iota-testnet", "iota-localnet":
		err = iota(ctx, *network, *action, *reference, hashes)
	default:
		err = errors.New("unknown benchmark network")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func ethereum(ctx context.Context, chainID int64, action string, hashes [3][32]byte) error {
	rpc, err := ethclient.DialContext(ctx, os.Getenv("BASELINE_ETHEREUM_RPC_URL"))
	if err != nil {
		return err
	}
	defer rpc.Close()
	id, err := rpc.ChainID(ctx)
	if err != nil {
		return err
	}
	if !id.IsInt64() || id.Int64() != chainID {
		return fmt.Errorf("unexpected chain ID %s", id)
	}
	var contract *registry.BaselineRegistry
	if action != "deploy" {
		value := os.Getenv("BASELINE_ETHEREUM_CONTRACT")
		if !common.IsHexAddress(value) {
			return errors.New("BASELINE_ETHEREUM_CONTRACT must contain the deployed address")
		}
		address := common.HexToAddress(value)
		contract, err = registry.NewBaselineRegistry(address, rpc)
		if err != nil {
			return err
		}
	}
	if action == "verify" {
		found, err := contract.HasHashes(&bind.CallOpts{Context: ctx}, hashes[0], hashes[1], hashes[2])
		if err != nil {
			return err
		}
		return output(map[string]bool{"found": found})
	}
	key, err := crypto.HexToECDSA(strings.TrimPrefix(os.Getenv("BASELINE_ETHEREUM_PRIVATE_KEY"), "0x"))
	if err != nil {
		return errors.New("BASELINE_ETHEREUM_PRIVATE_KEY must contain a signing key")
	}
	auth, err := bind.NewKeyedTransactorWithChainID(key, id)
	if err != nil {
		return err
	}
	auth.Context = ctx
	started := time.Now().UTC()
	var tx *types.Transaction
	var address common.Address
	if action == "deploy" {
		writer := os.Getenv("BASELINE_ETHEREUM_WRITER")
		if !common.IsHexAddress(writer) {
			return errors.New("BASELINE_ETHEREUM_WRITER must contain the oracle address")
		}
		address, tx, _, err = registry.DeployBaselineRegistry(auth, rpc, common.HexToAddress(writer))
	} else {
		tx, err = contract.AddHashes(auth, hashes[0], hashes[1], hashes[2])
	}
	if err != nil {
		return err
	}
	receipt, err := bind.WaitMined(ctx, rpc, tx)
	if err != nil {
		return fmt.Errorf("check transaction %s: %w", tx.Hash(), err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("transaction %s reverted", tx.Hash())
	}
	result := map[string]any{
		"txRef": tx.Hash(), "submittedAt": started, "includedAt": time.Now().UTC(),
		"gasUsed": receipt.GasUsed, "transactionSize": tx.Size(), "chainId": chainID,
	}
	if action == "deploy" {
		result["contractAddress"] = address
	}
	return output(result)
}

func iota(ctx context.Context, network, action, reference string, hashes [3][32]byte) error {
	if action == "deploy" {
		return errors.New("IOTA uses the existing notarization package")
	}
	db, err := store.NewSQLiteIndex(":memory:")
	if err != nil {
		return err
	}
	defer db.Close()
	adapter, err := chain.NewIOTAAdapter(ctx, os.Getenv("IOTA_SERVICE_URL"), os.Getenv("IOTA_SERVICE_TOKEN"),
		strings.Split(os.Getenv("IOTA_TRUSTED_SIGNERS"), ","), db)
	if err != nil {
		return err
	}
	if adapter.Name() != network {
		return fmt.Errorf("IOTA service uses %s", adapter.Name())
	}
	fingerprints := chain.BaselineHashes{
		ArtifactHash: "0x" + hex.EncodeToString(hashes[0][:]),
		SBOMHash:     "0x" + hex.EncodeToString(hashes[1][:]),
		ScanHash:     "0x" + hex.EncodeToString(hashes[2][:]),
	}
	if action == "verify" {
		found, err := adapter.VerifyBaseline(ctx, reference, fingerprints)
		if err != nil {
			return err
		}
		return output(map[string]bool{"found": found})
	}
	receipt, err := adapter.WriteBaseline(ctx, fingerprints)
	if err != nil {
		if receipt != nil {
			return fmt.Errorf("check IOTA digest %s: %w", receipt.TxRef, err)
		}
		return err
	}
	return output(receipt)
}

func output(value any) error {
	return json.NewEncoder(os.Stdout).Encode(value)
}
