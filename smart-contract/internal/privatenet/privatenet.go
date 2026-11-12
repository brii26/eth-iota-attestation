// Package privatenet holds the helpers shared by the private Geth commands.
package privatenet

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// ChainID is the developer chain ID that Geth --dev uses.
const ChainID = 1337

// TxTimeout bounds how long a command waits for one transaction.
const TxTimeout = 60 * time.Second

var privateKeyPattern = regexp.MustCompile(`^0x[0-9a-fA-F]{64}$`)

// LoadEnvFile sets variables from a KEY=VALUE file without overriding the
// environment. A missing file is not an error.
func LoadEnvFile(path string) error {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if _, set := os.LookupEnv(key); !set {
			if err := os.Setenv(key, value); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}

// Getenv returns the variable or the fallback when it is empty.
func Getenv(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

// DeploymentFile is where the public deployment receipt is stored.
func DeploymentFile() string {
	return Getenv("DEPLOYMENT_FILE", "deployments/ethereum-private.json")
}

// RPCURL is the private chain endpoint.
func RPCURL() string {
	return Getenv("ETHEREUM_RPC_URL", "http://127.0.0.1:8545")
}

// Account is a signing account loaded from the environment.
type Account struct {
	Key     *ecdsa.PrivateKey
	Address common.Address
}

// LoadAccount reads a 32-byte hex private key from the named variable.
func LoadAccount(name string) (Account, error) {
	value := os.Getenv(name)
	if !privateKeyPattern.MatchString(value) {
		return Account{}, fmt.Errorf("%s must contain a 32-byte hex private key", name)
	}
	key, err := crypto.HexToECDSA(strings.TrimPrefix(value, "0x"))
	if err != nil {
		return Account{}, fmt.Errorf("%s is not a valid private key", name)
	}
	return Account{Key: key, Address: crypto.PubkeyToAddress(key.PublicKey)}, nil
}

// Opts returns signing options for the account on the private chain.
func (a Account) Opts() (*bind.TransactOpts, error) {
	return bind.NewKeyedTransactorWithChainID(a.Key, big.NewInt(ChainID))
}

// Network is a connection that has been checked to be the developer chain.
type Network struct {
	RPC    *rpc.Client
	Client *ethclient.Client
}

// Dial connects to the private chain and refuses any other chain ID.
func Dial(ctx context.Context) (*Network, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	raw, err := rpc.DialContext(ctx, RPCURL())
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", RPCURL(), err)
	}
	client := ethclient.NewClient(raw)
	id, err := client.ChainID(ctx)
	if err != nil {
		raw.Close()
		return nil, fmt.Errorf("read chain ID: %w", err)
	}
	if id.Int64() != ChainID {
		raw.Close()
		return nil, fmt.Errorf("expected the private developer chain with chain ID %d, got %s", ChainID, id)
	}
	return &Network{RPC: raw, Client: client}, nil
}

// Close releases the connection.
func (n *Network) Close() { n.RPC.Close() }

// WaitSuccess waits for the transaction and requires a successful receipt.
func (n *Network) WaitSuccess(ctx context.Context, tx *types.Transaction, what string) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(ctx, TxTimeout)
	defer cancel()
	receipt, err := bind.WaitMined(ctx, n.Client, tx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", what, err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, fmt.Errorf("%s failed in transaction %s", what, tx.Hash())
	}
	return receipt, nil
}
