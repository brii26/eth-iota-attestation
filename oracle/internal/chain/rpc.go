package chain

import (
	"context"
	"errors"
	"fmt"
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
	rpcClient  *rpc.Client
	ethClient  *ethclient.Client
	maxRetries int
}

// NewMultiRPCClient creates an RPC client pool that tries endpoints in order.
func NewMultiRPCClient(ctx context.Context, endpoints []string) (*MultiRPCClient, error) {
	if len(endpoints) == 0 {
		return nil, errors.New("at least one RPC endpoint is required")
	}

	m := &MultiRPCClient{
		endpoints:  endpoints,
		maxRetries: len(endpoints) * 2,
	}

	if err := m.connect(ctx, 0); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *MultiRPCClient) connect(ctx context.Context, idx int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.rpcClient != nil {
		m.rpcClient.Close()
	}

	endpoint := m.endpoints[idx]
	raw, err := rpc.DialContext(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("dial rpc %s: %w", endpoint, err)
	}

	m.rpcClient = raw
	m.ethClient = ethclient.NewClient(raw)
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
	m.mu.Lock()
	nextIdx := (m.activeIdx + 1) % len(m.endpoints)
	m.mu.Unlock()

	return m.connect(ctx, nextIdx)
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
	if m.rpcClient != nil {
		m.rpcClient.Close()
	}
}
