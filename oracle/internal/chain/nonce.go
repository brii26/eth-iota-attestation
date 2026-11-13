package chain

import (
	"context"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// NonceManager serializes transaction signing for an Ethereum address to guarantee
// sequential, collision-free nonces under concurrent submissions.
type NonceManager struct {
	mu          sync.Mutex
	address     common.Address
	current     uint64
	initialized bool
}

// NewNonceManager creates a thread-safe nonce coordinator for the specified signing address.
func NewNonceManager(address common.Address) *NonceManager {
	return &NonceManager{address: address}
}

// Acquire reserves the next nonce. It returns the nonce and a completion callback
// that must be invoked with true if the transaction was broadcast successfully,
// or false if it failed and the nonce counter should be re-synced.
func (n *NonceManager) Acquire(ctx context.Context, client *ethclient.Client) (uint64, func(success bool), error) {
	n.mu.Lock()

	if !n.initialized {
		pendingNonce, err := client.PendingNonceAt(ctx, n.address)
		if err != nil {
			n.mu.Unlock()
			return 0, nil, fmt.Errorf("fetch pending nonce for %s: %w", n.address.Hex(), err)
		}
		n.current = pendingNonce
		n.initialized = true
	}

	allocatedNonce := n.current

	done := func(success bool) {
		if success {
			n.current++
		} else {
			// Re-sync with node on next transaction if submission failed
			n.initialized = false
		}
		n.mu.Unlock()
	}

	return allocatedNonce, done, nil
}
