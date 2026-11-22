package chain

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/brii26/eth-iota-attestation/oracle/internal/store"
	"github.com/brii26/eth-iota-attestation/smart-contract/registry"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestEthereumPendingReceiptSurvivesRestart(t *testing.T) {
	signingKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	contract := common.HexToAddress("0x1000000000000000000000000000000000000001")
	abi, err := registry.AttestationRegistryMetaData.GetAbi()
	if err != nil {
		t.Fatal(err)
	}
	event := abi.Events["RecordAdded"]
	args, err := event.Inputs.NonIndexed().Pack("app", "1.0.0", "first-scan")
	if err != nil {
		t.Fatal(err)
	}
	var broadcast atomic.Pointer[types.Transaction]
	var confirmed atomic.Bool
	var sends atomic.Int32
	service := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var call struct {
			ID     json.RawMessage   `json:"id"`
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&call); err != nil {
			t.Error(err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		var result any
		switch call.Method {
		case "eth_chainId":
			result = "0x539"
		case "eth_getTransactionCount":
			result = "0x0"
		case "eth_getBlockByNumber":
			result = &types.Header{Difficulty: new(big.Int), Number: big.NewInt(1), GasLimit: 30_000_000, BaseFee: big.NewInt(1)}
		case "eth_maxPriorityFeePerGas":
			result = "0x1"
		case "eth_getCode":
			result = "0x01"
		case "eth_estimateGas":
			result = "0x7a120"
		case "eth_getTransactionByHash":
		case "eth_sendRawTransaction":
			var signed string
			var tx types.Transaction
			if err := json.Unmarshal(call.Params[0], &signed); err != nil {
				t.Error(err)
			}
			if err := tx.UnmarshalBinary(common.FromHex(signed)); err != nil {
				t.Error(err)
			}
			sends.Add(1)
			broadcast.Store(&tx)
			result = tx.Hash()
		case "eth_getTransactionReceipt":
			if tx := broadcast.Load(); confirmed.Load() && tx != nil {
				result = &types.Receipt{
					Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(2), TxHash: tx.Hash(),
					Logs: []*types.Log{{
						Address: contract, Topics: []common.Hash{event.ID, common.BigToHash(big.NewInt(1)), {}, {}},
						Data: args, TxHash: tx.Hash(), BlockNumber: 2,
					}},
				}
			}
		default:
			t.Errorf("unexpected Ethereum RPC call: %s", call.Method)
			http.Error(w, "not supported", http.StatusNotImplemented)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": call.ID, "result": result})
	}))
	defer service.Close()

	path := filepath.Join(t.TempDir(), "index.sqlite")
	index, err := store.NewSQLiteIndex(path)
	if err != nil {
		t.Fatal(err)
	}
	const network, scanID = "ethereum-private", "first-scan"
	if err := index.Save(t.Context(), store.AttestationRecord{
		Network: network, App: "app", Version: "1.0.0", ScanID: scanID, Status: "pending", SubmittedAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	newAdapter := func(idx *store.SQLiteIndex) *EthereumAdapter {
		t.Helper()
		adapter, err := NewEthereumAdapter(t.Context(), EthereumConfig{
			Name: network, Endpoints: []string{service.URL}, ContractAddress: contract.Hex(),
			PrivateKeyHex: common.Bytes2Hex(crypto.FromECDSA(signingKey)), Index: idx, TxTimeout: 250 * time.Millisecond,
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(adapter.multiRPC.Close)
		return adapter
	}
	adapter := newAdapter(index)
	_, err = adapter.WriteRecord(t.Context(), Record{
		Network: network, App: "app", Version: "1.0.0", ScanID: scanID,
		ArtifactHash: [32]byte{1}, SBOMHash: [32]byte{2}, ScanHash: [32]byte{3},
	})
	if !errors.Is(err, ErrTransactionPending) || sends.Load() != 1 {
		t.Fatalf("want one broadcast and pending outcome, got %d, %v", sends.Load(), err)
	}
	status, err := index.GetStatus(t.Context(), network, scanID)
	if err != nil || status == nil || status.Status != "pending" || status.TxRef != broadcast.Load().Hash().Hex() {
		t.Fatalf("missing persisted pending transaction: %+v, %v", status, err)
	}
	if err := index.Close(); err != nil {
		t.Fatal(err)
	}
	index, err = store.NewSQLiteIndex(path)
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()
	confirmed.Store(true)
	adapter = newAdapter(index)
	if err := adapter.Reconcile(t.Context()); err != nil {
		t.Fatal(err)
	}
	status, err = index.GetStatus(t.Context(), network, scanID)
	if err != nil || status == nil || status.Status != "included" || status.RecordID != "1" || sends.Load() != 1 {
		t.Fatalf("recovery did not reuse the original transaction: %+v, sends=%d, %v", status, sends.Load(), err)
	}
}

func TestEthereumFallbackRejectsWrongChain(t *testing.T) {
	endpoint := func(chain string) *httptest.Server {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var call struct{ ID json.RawMessage }
			_ = json.NewDecoder(r.Body).Decode(&call)
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": call.ID, "result": chain})
		}))
		t.Cleanup(server.Close)
		return server
	}
	primary, wrong, backup := endpoint("0x539"), endpoint("0x1"), endpoint("0x539")
	pool, err := NewMultiRPCClient(context.Background(), []string{primary.URL, wrong.URL, backup.URL}, big.NewInt(1337))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Failover(t.Context()); err != nil || pool.CurrentEndpoint() != backup.URL {
		t.Fatalf("fallback picked wrong network: %s, %v", pool.CurrentEndpoint(), err)
	}
}
