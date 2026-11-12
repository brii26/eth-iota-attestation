// Command deploy-private deploys AttestationRegistry to the private Geth chain,
// authorizes the oracle, writes one smoke record, and saves a public receipt.
package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/brii26/eth-iota-attestation/smart-contract/internal/privatenet"
	"github.com/brii26/eth-iota-attestation/smart-contract/registry"
)

var oneEther = big.NewInt(1e18)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "deploy-private:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	if err := privatenet.LoadEnvFile(".env"); err != nil {
		return err
	}
	admin, err := privatenet.LoadAccount("ETHEREUM_ADMIN_PRIVATE_KEY")
	if err != nil {
		return err
	}
	oracle, err := privatenet.LoadAccount("ETHEREUM_ORACLE_PRIVATE_KEY")
	if err != nil {
		return err
	}
	if admin.Address == oracle.Address {
		return fmt.Errorf("admin and oracle must use different accounts")
	}
	network, err := privatenet.Dial(ctx)
	if err != nil {
		return err
	}
	defer network.Close()

	if os.Getenv("FUND_DEV_ACCOUNTS") == "1" {
		if err := fund(ctx, network, admin.Address, oracle.Address); err != nil {
			return err
		}
	}

	adminOpts, err := admin.Opts()
	if err != nil {
		return err
	}
	address, deployTx, contract, err := registry.DeployAttestationRegistry(adminOpts, network.Client)
	if err != nil {
		return fmt.Errorf("deploy: %w", err)
	}
	if _, err := network.WaitSuccess(ctx, deployTx, "deployment"); err != nil {
		return err
	}

	adminOpts, _ = admin.Opts()
	permissionTx, err := contract.AuthorizeOracle(adminOpts, oracle.Address)
	if err != nil {
		return fmt.Errorf("authorize oracle: %w", err)
	}
	if _, err := network.WaitSuccess(ctx, permissionTx, "oracle authorization"); err != nil {
		return err
	}

	scanID, err := newUUID()
	if err != nil {
		return err
	}
	const app, version = "private-chain-smoke", "1.0.0"
	artifactHash := sha256.Sum256([]byte("demo artifact bytes"))
	sbomHash := sha256.Sum256([]byte("demo SBOM bytes"))
	scanHash := sha256.Sum256([]byte("demo scan bytes"))

	oracleOpts, err := oracle.Opts()
	if err != nil {
		return err
	}
	recordTx, err := contract.AddRecord(oracleOpts, app, version, scanID, artifactHash, sbomHash, scanHash)
	if err != nil {
		return fmt.Errorf("add record: %w", err)
	}
	recordReceipt, err := network.WaitSuccess(ctx, recordTx, "smoke record transaction")
	if err != nil {
		return err
	}

	latest, err := contract.GetLatestRecord(nil, app, version)
	if err != nil {
		return fmt.Errorf("read latest record: %w", err)
	}
	if latest.ScanId != scanID || latest.Signer != oracle.Address {
		return fmt.Errorf("stored record does not match the smoke record")
	}
	recordID, err := contract.RecordCount(nil)
	if err != nil {
		return err
	}

	receipt := map[string]any{
		"chainId":         privatenet.ChainID,
		"contractAddress": address.Hex(),
		"admin":           admin.Address.Hex(),
		"oracle":          oracle.Address.Hex(),
		"deploymentTx":    deployTx.Hash().Hex(),
		"permissionTx":    permissionTx.Hash().Hex(),
		"recordTx":        recordTx.Hash().Hex(),
		"recordBlockHash": recordReceipt.BlockHash.Hex(),
		"recordId":        recordID.String(),
		"app":             app,
		"version":         version,
		"scanId":          scanID,
		"artifactHash":    hexutil.Encode(artifactHash[:]),
		"sbomHash":        hexutil.Encode(sbomHash[:]),
		"scanHash":        hexutil.Encode(scanHash[:]),
		"timestamp":       latest.Timestamp.String(),
	}
	out, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	file := privatenet.DeploymentFile()
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(file, out, 0o600); err != nil {
		return err
	}
	fmt.Print(string(out))
	return nil
}

// fund sends ether from the unlocked developer account to the given accounts
// when their balance is below one ether.
func fund(ctx context.Context, network *privatenet.Network, targets ...common.Address) error {
	var accounts []common.Address
	if err := network.RPC.CallContext(ctx, &accounts, "eth_accounts"); err != nil {
		return fmt.Errorf("list developer accounts: %w", err)
	}
	if len(accounts) == 0 {
		return fmt.Errorf("no unlocked developer account is available for funding")
	}
	for _, target := range targets {
		balance, err := network.Client.BalanceAt(ctx, target, nil)
		if err != nil {
			return err
		}
		if balance.Cmp(oneEther) >= 0 {
			continue
		}
		var hash common.Hash
		err = network.RPC.CallContext(ctx, &hash, "eth_sendTransaction", map[string]any{
			"from":  accounts[0],
			"to":    target,
			"value": hexutil.EncodeBig(new(big.Int).Mul(big.NewInt(10), oneEther)),
		})
		if err != nil {
			return fmt.Errorf("fund %s: %w", target, err)
		}
		if err := waitHash(ctx, network, hash); err != nil {
			return fmt.Errorf("fund %s: %w", target, err)
		}
	}
	return nil
}

func waitHash(ctx context.Context, network *privatenet.Network, hash common.Hash) error {
	ctx, cancel := context.WithTimeout(ctx, privatenet.TxTimeout)
	defer cancel()
	for {
		receipt, err := network.Client.TransactionReceipt(ctx, hash)
		if err == nil {
			if receipt.Status != types.ReceiptStatusSuccessful {
				return fmt.Errorf("developer account funding failed")
			}
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func newUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
