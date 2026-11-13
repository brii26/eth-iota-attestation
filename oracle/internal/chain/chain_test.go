package chain

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func TestRouter(t *testing.T) {
	router := NewRouter()
	mock := NewMockAdapter("mock")

	router.Register(mock)

	// Registered lookup
	ad, err := router.Get("mock")
	if err != nil || ad == nil {
		t.Fatalf("expected to resolve mock adapter, got %v", err)
	}

	// Unregistered lookup
	_, err = router.Get("unregistered-chain")
	if err == nil {
		t.Fatalf("expected error for unregistered network, got nil")
	}
}

func TestMultiRPCClientFailover(t *testing.T) {
	ctx := context.Background()
	primaryHit := false
	backupHit := false

	// Primary server fails
	primaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		primaryHit = true
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer primaryServer.Close()

	// Backup server succeeds (mock eth_chainId jsonrpc response)
	backupServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backupHit = true
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x539"}`))
	}))
	defer backupServer.Close()

	client, err := NewMultiRPCClient(ctx, []string{primaryServer.URL, backupServer.URL})
	if err != nil {
		t.Fatalf("failed to create multi-rpc client: %v", err)
	}
	defer client.Close()

	err = client.ExecuteWithRetry(ctx, func(c *ethclient.Client) error {
		_, err := c.ChainID(ctx)
		return err
	})
	if err != nil {
		t.Fatalf("expected retry/failover to succeed, got %v", err)
	}

	if !primaryHit || !backupHit {
		t.Errorf("expected both primary and backup to be hit, got primary=%v, backup=%v", primaryHit, backupHit)
	}
}

func TestNonceManagerConcurrency(t *testing.T) {
	ctx := context.Background()
	// Mock JSON-RPC server answering eth_getTransactionCount
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x10"}`)) // initial nonce 16
	}))
	defer server.Close()

	client, err := ethclient.Dial(server.URL)
	if err != nil {
		t.Fatalf("failed to dial mock rpc: %v", err)
	}
	defer client.Close()

	addr := common.HexToAddress("0x1111111111111111111111111111111111111111")
	mgr := NewNonceManager(addr)

	var wg sync.WaitGroup
	count := 20
	nonces := make([]uint64, count)
	var mu sync.Mutex

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			n, done, err := mgr.Acquire(ctx, client)
			if err != nil {
				t.Errorf("failed to acquire nonce: %v", err)
				return
			}
			time.Sleep(2 * time.Millisecond) // simulate work
			done(true)

			mu.Lock()
			nonces[idx] = n
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify all nonces are distinct and sequential from 16 to 16 + count - 1
	seen := make(map[uint64]bool)
	for _, n := range nonces {
		if seen[n] {
			t.Errorf("duplicate nonce detected: %d", n)
		}
		seen[n] = true
		if n < 16 || n >= 16+uint64(count) {
			t.Errorf("nonce out of expected range: %d", n)
		}
	}
	if len(seen) != count {
		t.Errorf("expected %d unique nonces, got %d", count, len(seen))
	}
}
