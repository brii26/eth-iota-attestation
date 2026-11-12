// Command verify-private checks a saved deployment receipt against the private
// chain. It needs no signing key.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/brii26/eth-iota-attestation/smart-contract/internal/privatenet"
	"github.com/brii26/eth-iota-attestation/smart-contract/registry"
)

type saved struct {
	ChainID         int    `json:"chainId"`
	ContractAddress string `json:"contractAddress"`
	Admin           string `json:"admin"`
	Oracle          string `json:"oracle"`
	RecordTx        string `json:"recordTx"`
	RecordBlockHash string `json:"recordBlockHash"`
	RecordID        string `json:"recordId"`
	App             string `json:"app"`
	Version         string `json:"version"`
	ScanID          string `json:"scanId"`
	ArtifactHash    string `json:"artifactHash"`
	SbomHash        string `json:"sbomHash"`
	ScanHash        string `json:"scanHash"`
	Timestamp       string `json:"timestamp"`
}

var recordIDPattern = regexp.MustCompile(`^[1-9][0-9]*$`)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "verify-private:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	if err := privatenet.LoadEnvFile(".env"); err != nil {
		return err
	}
	raw, err := os.ReadFile(privatenet.DeploymentFile())
	if err != nil {
		return err
	}
	var s saved
	if err := json.Unmarshal(raw, &s); err != nil {
		return fmt.Errorf("parse deployment receipt: %w", err)
	}
	if s.ChainID != privatenet.ChainID {
		return fmt.Errorf("deployment receipt must belong to chain ID %d", privatenet.ChainID)
	}
	if !common.IsHexAddress(s.ContractAddress) {
		return fmt.Errorf("receipt has no valid contract address")
	}
	if !recordIDPattern.MatchString(s.RecordID) {
		return fmt.Errorf("receipt has no valid record ID")
	}
	if _, err := hexutil.Decode(s.RecordTx); err != nil {
		return fmt.Errorf("receipt has no valid record transaction")
	}

	network, err := privatenet.Dial(ctx)
	if err != nil {
		return err
	}
	defer network.Close()
	contract, err := registry.NewAttestationRegistry(common.HexToAddress(s.ContractAddress), network.Client)
	if err != nil {
		return err
	}
	receipt, err := network.Client.TransactionReceipt(ctx, common.HexToHash(s.RecordTx))
	if err != nil {
		return fmt.Errorf("read record transaction: %w", err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("record transaction did not succeed")
	}
	if !strings.EqualFold(receipt.BlockHash.Hex(), s.RecordBlockHash) {
		return fmt.Errorf("recorded transaction changed chains or blocks")
	}

	id, _ := new(big.Int).SetString(s.RecordID, 10)
	record, err := contract.GetRecord(nil, id)
	if err != nil {
		return fmt.Errorf("read record: %w", err)
	}
	checks := []struct {
		name      string
		got, want string
	}{
		{"app", record.App, s.App},
		{"version", record.Version, s.Version},
		{"scanId", record.ScanId, s.ScanID},
		{"artifactHash", hexutil.Encode(record.ArtifactHash[:]), s.ArtifactHash},
		{"sbomHash", hexutil.Encode(record.SbomHash[:]), s.SbomHash},
		{"scanHash", hexutil.Encode(record.ScanHash[:]), s.ScanHash},
		{"timestamp", record.Timestamp.String(), s.Timestamp},
	}
	for _, c := range checks {
		if c.got != c.want {
			return fmt.Errorf("stored %s differs from deployment receipt", c.name)
		}
	}
	if !strings.EqualFold(record.Signer.Hex(), s.Oracle) {
		return fmt.Errorf("stored signer differs from deployment receipt")
	}
	admin, err := contract.Admin(nil)
	if err != nil {
		return err
	}
	if !strings.EqualFold(admin.Hex(), s.Admin) {
		return fmt.Errorf("stored administrator differs from deployment receipt")
	}

	out, err := json.MarshalIndent(map[string]any{
		"status":          "verified",
		"chainId":         privatenet.ChainID,
		"contractAddress": s.ContractAddress,
		"recordId":        s.RecordID,
		"scanId":          record.ScanId,
		"blockHash":       receipt.BlockHash.Hex(),
	}, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}
