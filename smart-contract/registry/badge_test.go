package registry_test

import (
	"math/big"
	"testing"

	"github.com/brii26/eth-iota-attestation/smart-contract/registry"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestAttestationBadgePermissions(t *testing.T) {
	f := newFixture(t)
	_, deployment, badge, err := registry.DeployAttestationBadge(f.opts(f.admin), f.client, f.oracle.addr)
	if err != nil {
		t.Fatal(err)
	}
	f.mine(deployment)
	f.abi, err = registry.AttestationBadgeMetaData.GetAbi()
	if err != nil {
		t.Fatal(err)
	}
	recordID := big.NewInt(17)
	_, err = badge.Mint(f.opts(f.stranger), f.stranger.addr, recordID)
	f.expectRevert(err, "UnauthorizedMinter", f.stranger.addr)
	f.send(f.oracle, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return badge.Mint(opts, f.stranger.addr, recordID)
	})
	tokenID, err := badge.BadgeOf(nil, f.stranger.addr, recordID)
	if err != nil || tokenID.Sign() == 0 {
		t.Fatalf("expected a badge for the recipient and record: %v, %v", tokenID, err)
	}
	found, err := badge.RecordOf(nil, tokenID)
	if err != nil || found.Cmp(recordID) != 0 {
		t.Fatalf("token links to record %v: %v", found, err)
	}
	_, err = badge.TransferFrom(f.opts(f.stranger), f.stranger.addr, f.admin.addr, tokenID)
	f.expectRevert(err, "BadgeNonTransferable")
	_, err = badge.Mint(f.opts(f.oracle), f.stranger.addr, recordID)
	f.expectRevert(err, "BadgeAlreadyMinted", f.stranger.addr, recordID)
	owned, err := badge.TokenOfOwnerByIndex(nil, f.stranger.addr, big.NewInt(0))
	if err != nil || owned.Cmp(tokenID) != 0 {
		t.Fatalf("wallet index returned token %v: %v", owned, err)
	}
}
