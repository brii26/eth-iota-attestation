package chain

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/brii26/eth-iota-attestation/smart-contract/registry"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type BadgeMinter func(context.Context, string, *big.Int) error

func (a *EthereumAdapter) BadgeMinter(ctx context.Context, contractAddress string) (BadgeMinter, error) {
	address := common.HexToAddress(contractAddress)
	if !common.IsHexAddress(contractAddress) || address == (common.Address{}) {
		return nil, errors.New("invalid badge contract address")
	}
	contract, err := registry.NewAttestationBadge(address, a.multiRPC.CurrentClient())
	if err != nil {
		return nil, err
	}
	minter, err := contract.Minter(&bind.CallOpts{Context: ctx})
	if err != nil {
		return nil, err
	}
	if minter != a.oracleAddr {
		return nil, errors.New("badge minter must be the configured oracle")
	}
	return func(ctx context.Context, recipient string, recordID *big.Int) error {
		if !common.IsHexAddress(recipient) || recordID == nil || recordID.Sign() <= 0 {
			return errors.New("invalid badge recipient or record ID")
		}
		to := common.HexToAddress(recipient)
		client := a.multiRPC.CurrentClient()
		contract, err := registry.NewAttestationBadge(address, client)
		if err != nil {
			return err
		}
		token, err := contract.BadgeOf(&bind.CallOpts{Context: ctx}, to, recordID)
		if err != nil || token.Sign() != 0 {
			return err
		}
		nonce, releaseNonce, err := a.nonceMgr.Acquire(ctx, client)
		if err != nil {
			return err
		}
		opts, err := bind.NewKeyedTransactorWithChainID(a.privKey, a.chainID)
		if err != nil {
			releaseNonce(false)
			return err
		}
		opts.Nonce = new(big.Int).SetUint64(nonce)
		opts.Context = ctx
		tx, err := contract.Mint(opts, to, recordID)
		releaseNonce(err == nil)
		if err != nil {
			return err
		}
		wait, cancel := context.WithTimeout(ctx, a.txTimeout)
		defer cancel()
		receipt, err := bind.WaitMined(wait, client, tx)
		if err != nil {
			return fmt.Errorf("wait for badge transaction %s: %w", tx.Hash(), err)
		}
		if receipt.Status != types.ReceiptStatusSuccessful {
			return fmt.Errorf("badge transaction %s reverted", tx.Hash())
		}
		return nil
	}, nil
}
