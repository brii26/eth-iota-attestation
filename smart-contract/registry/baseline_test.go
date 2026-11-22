package registry

import (
	"crypto/sha256"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
)

func TestBaselineMatchesFilesWithoutReleaseIdentity(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(1337))
	if err != nil {
		t.Fatal(err)
	}
	chain := simulated.NewBackend(types.GenesisAlloc{
		auth.From: {Balance: new(big.Int).SetUint64(5_000_000_000_000_000_000)},
	})
	defer chain.Close()
	_, _, baseline, err := DeployBaselineRegistry(auth, chain.Client(), auth.From)
	if err != nil {
		t.Fatal(err)
	}
	chain.Commit()

	artifact := sha256.Sum256([]byte("package"))
	sbom := sha256.Sum256([]byte("sbom"))
	scan := sha256.Sum256([]byte("scan"))
	matches := func(artifact [32]byte) bool {
		t.Helper()
		ok, err := baseline.HasHashes(&bind.CallOpts{Context: t.Context()}, artifact, sbom, scan)
		if err != nil {
			t.Fatal(err)
		}
		return ok
	}
	if matches(artifact) {
		t.Fatal("unwritten files matched")
	}
	for range 2 {
		tx, err := baseline.AddHashes(auth, artifact, sbom, scan)
		if err != nil {
			t.Fatal(err)
		}
		chain.Commit()
		receipt, err := chain.Client().TransactionReceipt(t.Context(), tx.Hash())
		if err != nil || receipt.Status != types.ReceiptStatusSuccessful {
			t.Fatalf("record transaction failed: %+v, %v", receipt, err)
		}
	}
	if !matches(artifact) || matches(common.Hash{1}) {
		t.Fatal("the baseline must match only the recorded file fingerprints")
	}
}
