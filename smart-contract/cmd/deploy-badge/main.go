package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/brii26/eth-iota-attestation/smart-contract/registry"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	minter := flag.String("minter", "", "Address of the oracle authorized to mint badges")
	chainID := flag.Uint64("chain-id", 1337, "Expected chain ID")
	flag.Parse()
	if err := deploy(*minter, *chainID); err != nil {
		log.Fatal(err)
	}
}

func deploy(minter string, expectedChainID uint64) error {
	address := common.HexToAddress(minter)
	if !common.IsHexAddress(minter) || address == (common.Address{}) {
		return errors.New("a valid nonzero oracle address is required")
	}
	rpc := os.Getenv("ETHEREUM_BADGE_RPC_URL")
	if rpc == "" {
		return errors.New("ETHEREUM_BADGE_RPC_URL is required")
	}
	key, err := crypto.HexToECDSA(strings.TrimPrefix(os.Getenv("ETHEREUM_BADGE_DEPLOYER_PRIVATE_KEY"), "0x"))
	if err != nil {
		return errors.New("a valid badge deployer private key is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	client, err := ethclient.DialContext(ctx, rpc)
	if err != nil {
		return err
	}
	defer client.Close()
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return err
	}
	if !chainID.IsUint64() || chainID.Uint64() != expectedChainID {
		return fmt.Errorf("RPC chain ID %s differs from the expected chain ID %d", chainID, expectedChainID)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		return err
	}
	auth.Context = ctx
	contract, tx, _, err := registry.DeployAttestationBadge(auth, client, address)
	if err != nil {
		return err
	}
	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		return fmt.Errorf("wait for badge deployment %s: %w", tx.Hash(), err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("badge deployment %s reverted", tx.Hash())
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{
		"chainId": chainID, "contractAddress": contract, "minter": address, "transaction": tx.Hash(),
	})
}
