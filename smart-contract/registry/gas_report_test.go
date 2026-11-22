package registry_test

import (
	"math/big"
	"testing"

	"github.com/brii26/eth-iota-attestation/smart-contract/registry"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestRegistryGasReport(t *testing.T) {
	fixture := newFixture(t)
	measuredOracle := common.HexToAddress("0x1111111111111111111111111111111111111111")
	authorized := fixture.send(fixture.admin, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return fixture.contract.AuthorizeOracle(opts, measuredOracle)
	})
	revoked := fixture.send(fixture.admin, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return fixture.contract.RevokeOracle(opts, measuredOracle)
	})
	recorded := fixture.add(fixture.oracle, defaultRecord())
	requested := fixture.send(fixture.stranger, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return fixture.contract.RequestVerification(opts, big.NewInt(1))
	})
	fulfilled := fixture.send(fixture.oracle, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return fixture.contract.FulfillVerification(opts, big.NewInt(1), registry.AttestationRegistryVerificationResult{Status: 1})
	})
	for _, method := range []struct {
		name    string
		receipt *types.Receipt
	}{{"addRecord", recorded}, {"authorizeOracle", authorized}, {"revokeOracle", revoked},
		{"requestVerification", requested}, {"fulfillVerification", fulfilled}} {
		t.Logf("%s=%d", method.name, method.receipt.GasUsed)
	}
}
