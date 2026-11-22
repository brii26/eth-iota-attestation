package chain

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/brii26/eth-iota-attestation/oracle/internal/store"
)

func TestIOTAResumeAndReconcile(t *testing.T) {
	ctx := t.Context()
	indexPath := filepath.Join(t.TempDir(), "index.sqlite")
	index, err := store.NewSQLiteIndex(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	signer := "0x" + strings.Repeat("1", 64)
	object := "0x" + strings.Repeat("2", 64)
	otherObject := "0x" + strings.Repeat("3", 64)
	hash := strings.Repeat("0", 64)
	serviceToken := strings.Repeat("t", 32)
	now := time.Now().UTC().Truncate(time.Second)
	prepared := store.IOTATransaction{Digest: "test-digest", Bytes: "signed-bytes", Signature: "signature"}
	var prepares, submits atomic.Int32
	service := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+serviceToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		respond := func(value any) { _ = json.NewEncoder(w).Encode(value) }
		switch r.URL.Path {
		case "/network":
			respond(map[string]string{"name": "iota-testnet", "signer": signer})
		case "/prepare":
			prepares.Add(1)
			respond(prepared)
		case "/submit":
			var got store.IOTATransaction
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil || got != prepared {
				t.Errorf("submit did not reuse signed transaction: %+v, %v", got, err)
				http.Error(w, "invalid", http.StatusBadRequest)
				return
			}
			if submits.Add(1) == 1 {
				http.Error(w, "simulated timeout", http.StatusBadGateway)
				return
			}
			respond(store.IOTAReceipt{TxRef: prepared.Digest, RecordID: object, Signer: signer, IncludedAt: &now})
		case "/records":
			respond(map[string][]string{"ids": {object, otherObject}})
		case "/record/" + object:
			respond(iotaRecord{
				iotaPayload: iotaPayload{1, "fixture", "1.0.0", "first", "0x" + hash, "0x" + hash, "0x" + hash},
				IOTAReceipt: store.IOTAReceipt{TxRef: prepared.Digest, RecordID: object, Signer: signer, IncludedAt: &now},
				Timestamp:   now,
			})
		case "/record/" + otherObject:
			later := now.Add(time.Second)
			respond(iotaRecord{
				iotaPayload: iotaPayload{1, "fixture", "1.0.0", "second", "0x" + hash, "0x" + hash, "0x" + hash},
				IOTAReceipt: store.IOTAReceipt{TxRef: "another-digest", RecordID: otherObject, Signer: signer, IncludedAt: &later},
				Timestamp:   later,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer service.Close()
	openAdapter := func(index *store.SQLiteIndex) *IOTAAdapter {
		t.Helper()
		adapter, err := NewIOTAAdapter(ctx, service.URL, serviceToken, []string{signer}, index)
		if err != nil {
			t.Fatal(err)
		}
		return adapter
	}
	if err := index.Save(ctx, store.AttestationRecord{
		Network: "iota-testnet", App: "fixture", Version: "1.0.0", ScanID: "first",
		ArtifactHash: hash, SBOMHash: hash, ScanHash: hash, Status: "pending", SubmittedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	adapter := openAdapter(index)
	if _, err := adapter.WriteRecord(ctx, Record{Network: adapter.Name(), App: "fixture", Version: "1.0.0", ScanID: "first"}); !errors.Is(err, ErrTransactionPending) {
		t.Fatalf("want pending, got %v", err)
	}
	if err := index.Close(); err != nil {
		t.Fatal(err)
	}
	index, err = store.NewSQLiteIndex(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()
	adapter = openAdapter(index)
	if err := adapter.Reconcile(ctx); err != nil {
		t.Fatal(err)
	}
	if prepares.Load() != 1 || submits.Load() != 2 {
		t.Fatalf("prepare calls %d, submit calls %d", prepares.Load(), submits.Load())
	}
	status, err := index.GetStatus(ctx, adapter.Name(), "first")
	if err != nil || status == nil || status.RecordID != object || status.Status != "included" {
		t.Fatalf("unexpected status: %+v, %v", status, err)
	}
	for _, query := range []struct {
		query Query
		want  string
	}{{Query{App: "fixture", Version: "1.0.0"}, otherObject}, {Query{ExternalRecordID: object}, object}} {
		record, err := adapter.ReadRecord(context.Background(), query.query)
		if err != nil || record.RecordID != query.want {
			t.Fatalf("unexpected record: %+v, %v", record, err)
		}
	}
}
