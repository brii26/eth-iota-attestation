package listener

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/brii26/eth-iota-attestation/oracle/internal/chain"
	"github.com/brii26/eth-iota-attestation/oracle/internal/verify"
	"github.com/brii26/eth-iota-attestation/smart-contract/registry"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func TestLocalBadgeRoundTrip(t *testing.T) {
	rpc := os.Getenv("ATTESTATION_BADGE_SMOKE_RPC")
	if rpc == "" {
		t.Skip("set ATTESTATION_BADGE_SMOKE_RPC to test badges on local Geth")
	}
	ctx := t.Context()
	client, err := ethclient.DialContext(ctx, rpc)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	id, err := client.ChainID(ctx)
	if err != nil || id.Cmp(big.NewInt(1337)) != 0 {
		t.Fatalf("badge smoke requires private Geth: %v, %v", id, err)
	}
	opts := func(variable string) *bind.TransactOpts {
		t.Helper()
		key, err := crypto.HexToECDSA(strings.TrimPrefix(os.Getenv(variable), "0x"))
		if err != nil {
			t.Fatalf("%s is required", variable)
		}
		auth, err := bind.NewKeyedTransactorWithChainID(key, id)
		if err != nil {
			t.Fatal(err)
		}
		auth.Context = ctx
		return auth
	}
	admin := opts("ETHEREUM_ADMIN_PRIVATE_KEY")
	oracle := opts("ETHEREUM_ORACLE_PRIVATE_KEY")
	mined := func(tx *types.Transaction, err error) *types.Receipt {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		receipt, err := bind.WaitMined(ctx, client, tx)
		if err != nil || receipt.Status != types.ReceiptStatusSuccessful {
			t.Fatalf("transaction %s failed: %v, %v", tx.Hash(), receipt, err)
		}
		return receipt
	}
	address, tx, contract, err := registry.DeployAttestationRegistry(admin, client)
	mined(tx, err)
	mined(contract.AuthorizeOracle(admin, oracle.From))
	badgeAddress, tx, badge, err := registry.DeployAttestationBadge(admin, client, oracle.From)
	mined(tx, err)

	adapter, err := chain.NewEthereumAdapter(ctx, chain.EthereumConfig{
		Endpoints: []string{rpc}, ContractAddress: address.Hex(), PrivateKeyHex: os.Getenv("ETHEREUM_ORACLE_PRIVATE_KEY"),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer adapter.Close()
	worker, _, index, blobs, fetcher := setupWorkerTest(t)
	defer index.Close()
	worker.cfg.Adapter = adapter
	worker.cfg.MintBadge, err = adapter.BadgeMinter(ctx, badgeAddress.Hex())
	if err != nil {
		t.Fatal(err)
	}
	artifact, sbom, scan := []byte("package"), []byte("sbom"), []byte("scan")
	artifactHash, _ := verify.ComputeSHA256(artifact)
	sbomHash, _ := verify.ComputeSHA256(sbom)
	scanHash, _ := verify.ComputeSHA256(scan)

	for _, test := range []struct {
		version   string
		integrity uint8
		badge     bool
	}{
		{"1.0.0", integrityValid, true},
		{"2.0.0", 2, false},
		{"3.0.0", integrityValid, false},
	} {
		name := "badge-smoke"
		key := fmt.Sprintf("%s/%s/%s/%s.json", adapter.Name(), name, test.version, test.version)
		blobs.sboms[key], blobs.scans[key] = sbom, scan
		fetcher.artifacts[name+"@"+test.version] = artifact
		if test.integrity != integrityValid {
			fetcher.artifacts[name+"@"+test.version] = []byte("changed")
		}
		if test.version == "3.0.0" {
			worker.cfg.MintBadge = func(context.Context, string, *big.Int) error {
				return errors.New("badge RPC unavailable")
			}
		}
		record, err := adapter.WriteRecord(ctx, chain.Record{
			Network: adapter.Name(), App: name, Version: test.version, ScanID: test.version,
			ArtifactHash: artifactHash, SBOMHash: sbomHash, ScanHash: scanHash,
		})
		if err != nil {
			t.Fatal(err)
		}
		recordID, ok := new(big.Int).SetString(record.RecordID, 10)
		if !ok {
			t.Fatalf("unexpected record ID %q", record.RecordID)
		}
		receipt := mined(contract.RequestVerification(admin, recordID))
		event, err := contract.ParseVerificationRequested(*receipt.Logs[0])
		if err != nil {
			t.Fatal(err)
		}
		processed, err := worker.PollOnce(ctx)
		if err != nil || processed != 1 {
			t.Fatalf("worker processed %d requests: %v", processed, err)
		}
		result, fulfilled, err := adapter.ReadVerificationResult(ctx, event.RequestId)
		if err != nil || !fulfilled || result.Status != test.integrity {
			t.Fatalf("verification result changed: %+v, fulfilled=%v, err=%v", result, fulfilled, err)
		}
		token, err := badge.BadgeOf(&bind.CallOpts{Context: ctx}, admin.From, recordID)
		if err != nil || (token.Sign() != 0) != test.badge {
			t.Fatalf("record %s expected badge=%v: %v, %v", recordID, test.badge, token, err)
		}
	}
}
