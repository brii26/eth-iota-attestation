package chain

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// MultiRPCClient wraps primary and fallback RPC endpoints with automatic failover and retry.
type MultiRPCClient struct {
	mu         sync.RWMutex
	endpoints  []string
	activeIdx  int
	clients    []*ethclient.Client
	chainID    *big.Int
	ethClient  *ethclient.Client
	maxRetries int
}

// NewMultiRPCClient creates an RPC client pool that tries endpoints in order.
func NewMultiRPCClient(ctx context.Context, endpoints []string, expectedChain ...*big.Int) (*MultiRPCClient, error) {
	if len(endpoints) == 0 {
		return nil, errors.New("at least one RPC endpoint is required")
	}

	m := &MultiRPCClient{
		endpoints:  endpoints,
		maxRetries: len(endpoints) * 2,
		clients:    make([]*ethclient.Client, len(endpoints)),
	}

	if len(expectedChain) > 0 {
		m.chainID = expectedChain[0]
	}
	var failures error
	for idx := range endpoints {
		if err := m.connect(ctx, idx); err == nil {
			return m, nil
		} else {
			failures = errors.Join(failures, err)
		}
	}
	return nil, failures
}

func (m *MultiRPCClient) connect(ctx context.Context, idx int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client := m.clients[idx]
	if client == nil {
		raw, err := rpc.DialContext(ctx, m.endpoints[idx])
		if err != nil {
			return fmt.Errorf("connect Ethereum RPC %d: %w", idx+1, err)
		}
		client = ethclient.NewClient(raw)
		if m.chainID != nil {
			probe, cancel := context.WithTimeout(ctx, 4*time.Second)
			id, err := client.ChainID(probe)
			cancel()
			if err != nil || id.Cmp(m.chainID) != 0 {
				client.Close()
				if err != nil {
					return fmt.Errorf("validate Ethereum RPC %d: %w", idx+1, err)
				}
				return fmt.Errorf("Ethereum RPC %d uses chain %s, expected %s", idx+1, id, m.chainID)
			}
		}
		m.clients[idx] = client
	}
	m.ethClient = client
	m.activeIdx = idx
	return nil
}

// CurrentClient returns the active ethclient.
func (m *MultiRPCClient) CurrentClient() *ethclient.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ethClient
}

// CurrentEndpoint returns the active RPC URL.
func (m *MultiRPCClient) CurrentEndpoint() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.endpoints[m.activeIdx]
}

// Failover switches to the next available RPC endpoint.
func (m *MultiRPCClient) Failover(ctx context.Context) error {
	m.mu.RLock()
	active := m.activeIdx
	m.mu.RUnlock()
	failures := errors.New("no Ethereum RPC fallback is available")
	for offset := 1; offset < len(m.endpoints); offset++ {
		if err := m.connect(ctx, (active+offset)%len(m.endpoints)); err == nil {
			return nil
		} else {
			failures = errors.Join(failures, err)
		}
	}
	return failures
}

// ExecuteWithRetry runs an action using the active client. If it fails with a network/RPC error,
// it fails over to the backup endpoint and retries with backoff.
func (m *MultiRPCClient) ExecuteWithRetry(ctx context.Context, action func(client *ethclient.Client) error) error {
	var lastErr error
	backoff := 50 * time.Millisecond

	for attempt := 0; attempt < m.maxRetries; attempt++ {
		client := m.CurrentClient()
		err := action(client)
		if err == nil {
			return nil
		}

		lastErr = err

		// Attempt failover if more than 1 endpoint
		if len(m.endpoints) > 1 {
			_ = m.Failover(ctx)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
			if backoff > 1*time.Second {
				backoff = 1 * time.Second
			}
		}
	}

	return fmt.Errorf("action failed after %d retries: %w", m.maxRetries, lastErr)
}

// Close closes the underlying active RPC connection.
func (m *MultiRPCClient) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, client := range m.clients {
		if client != nil {
			client.Close()
		}
	}
}
