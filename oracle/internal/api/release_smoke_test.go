package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/brii26/eth-iota-attestation/oracle/internal/chain"
	"github.com/brii26/eth-iota-attestation/oracle/internal/model"
	"github.com/brii26/eth-iota-attestation/oracle/internal/store"
	"github.com/brii26/eth-iota-attestation/oracle/internal/verify"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func TestLocalReleaseRoundTrip(t *testing.T) {
	directory := os.Getenv("ATTESTATION_SMOKE_FIXTURE_DIR")
	if directory == "" {
		t.Skip("set ATTESTATION_SMOKE_FIXTURE_DIR to verify a published release against local services")
	}
	app, version, archive := os.Getenv("ATTESTATION_SMOKE_APP"), os.Getenv("ATTESTATION_SMOKE_VERSION"), os.Getenv("ATTESTATION_SMOKE_ARCHIVE")
	if app == "" || version == "" || !filepath.IsLocal(archive) {
		t.Fatal("fixture app, version, and relative archive path are required")
	}
	read := func(path string) []byte {
		t.Helper()
		content, err := os.ReadFile(filepath.Join(directory, path))
		if err != nil {
			t.Fatal(err)
		}
		return content
	}
	sbom, scan := os.Getenv("ATTESTATION_SMOKE_SBOM"), os.Getenv("ATTESTATION_SMOKE_SCAN")
	if !filepath.IsLocal(sbom) || !filepath.IsLocal(scan) {
		t.Fatal("relative SBOM and scan paths are required")
	}
	files := [][]byte{read(archive), read(sbom), read(scan)}
	var hashes [3]string
	for index, file := range files {
		_, hashes[index] = verify.ComputeSHA256(file)
	}
	index, err := store.NewSQLiteIndex(filepath.Join(t.TempDir(), "release.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()
	ctx := t.Context()
	network := os.Getenv("ATTESTATION_SMOKE_NETWORK")
	var adapter chain.Adapter
	switch network {
	case "", "ethereum-private":
		adapter, err = chain.NewEthereumAdapter(ctx, chain.EthereumConfig{
			Name: "ethereum-private", Endpoints: []string{os.Getenv("ETHEREUM_PRIVATE_RPC_URL")},
			ContractAddress: os.Getenv("ETHEREUM_PRIVATE_CONTRACT_ADDRESS"),
			PrivateKeyHex:   os.Getenv("ETHEREUM_ORACLE_PRIVATE_KEY"), Index: index,
		})
	case "iota-localnet", "iota-testnet":
		adapter, err = chain.NewIOTAAdapter(ctx, os.Getenv("IOTA_SERVICE_URL"), os.Getenv("IOTA_SERVICE_TOKEN"),
			strings.Split(os.Getenv("IOTA_TRUSTED_SIGNERS"), ","), index)
	default:
		t.Fatalf("unsupported smoke network %q", network)
	}

	if err != nil {
		t.Fatal(err)
	}
	if network != "" && adapter.Name() != network {
		t.Fatalf("configuration resolved %s, expected %s", adapter.Name(), network)
	}
	blob, err := store.NewS3Store(ctx, store.S3Config{Endpoint: os.Getenv("RUSTFS_ENDPOINT"),
		AccessKey: os.Getenv("RUSTFS_ACCESS_KEY"), SecretKey: os.Getenv("RUSTFS_SECRET_KEY")})
	if err != nil {
		t.Fatal(err)
	}
	var id [16]byte
	rand.Read(id[:])
	scanID := fmt.Sprintf("release-smoke-%x", id)
	token := fmt.Sprintf("test-pipeline-%x", id)
	source := verify.NewVerdaccioClient(os.Getenv("VERDACCIO_URL"))
	config := Config{S3Store: blob, ChainAdapter: adapter, Index: index, VerdaccioClient: source,
		AuthorizePipeline: func(_ context.Context, candidate, name string) bool {
			return candidate == token && name == app
		}}
	ipfsURL := os.Getenv("ATTESTATION_SMOKE_IPFS_KUBO_URL")
	if ipfsURL != "" {
		config.IPFSSBOM, err = store.NewIPFSSBOMStore(blob, index, ipfsURL, app)
		if err != nil {
			t.Fatal(err)
		}
	}
	server := NewServer(config)
	post := func(path string, data any, want ...int) *httptest.ResponseRecorder {
		t.Helper()
		body, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}
		request := httptest.NewRequestWithContext(ctx, http.MethodPost, path, bytes.NewReader(body))
		request.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if !slices.Contains(want, response.Code) {
			t.Fatalf("%s got HTTP %d, wanted one of %v: %s", path, response.Code, want, response.Body.String())
		}
		return response
	}
	attestation := model.AttestationRequest{
		Network: adapter.Name(), App: app, Version: version, ScanID: scanID,
		ClaimedHashes: model.ClaimedHashes{Artifact: hashes[0], SBOM: hashes[1], Scan: hashes[2]},
		SBOMBytes:     files[1], ScanBytes: files[2],
	}
	response := post("/api/v1/attestations", attestation, http.StatusCreated, http.StatusAccepted)
	var attested model.AttestationStatus
	if err := json.Unmarshal(response.Body.Bytes(), &attested); err != nil || attested.TxRef == "" {
		t.Fatalf("missing transaction reference: %+v, %v", attested, err)
	}
	if attested.Status == "pending" {
		reconciler, ok := adapter.(interface{ Reconcile(context.Context) error })
		if !ok {
			t.Fatal("the pending transaction needs a reconciler")
		}
		wait, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		for attested.Status == "pending" && wait.Err() == nil {
			if err := reconciler.Reconcile(wait); err != nil {
				t.Fatal(err)
			}
			status, err := index.GetStatus(wait, adapter.Name(), scanID)
			if err != nil || status == nil || status.TxRef != attested.TxRef {
				t.Fatalf("transaction status could not be recovered: %+v, %v", status, err)
			}
			attested = *status
			if attested.Status == "pending" {
				select {
				case <-wait.Done():
				case <-time.After(time.Second):
				}
			}
		}
	}
	if attested.Status != "included" || attested.RecordID == "" {
		t.Fatalf("missing included record: %+v", attested)
	}
	deadline := time.Now().Add(30 * time.Second)
	var verified model.VerifyResponse
	for {
		response = post("/api/v1/verify", model.VerifyRequest{
			Network: adapter.Name(), App: app, Version: version, RecordID: attested.RecordID,
		}, http.StatusOK)
		if err := json.Unmarshal(response.Body.Bytes(), &verified); err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(adapter.Name(), "iota-") || verified.Status != model.StatusFileUnavailable ||
			time.Now().After(deadline) {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(time.Second):
		}
	}
	if verified.Status != model.StatusValid || verified.Stale || verified.GrypeDBBuiltAt == nil || verified.ScanID != scanID {
		t.Fatalf("release verification failed: %+v", verified)
	}
	t.Logf("verified %s@%s on %s in transaction %s", app, version, adapter.Name(), attested.TxRef)
	if ipfsURL == "" {
		return
	}

	endpoint, err := url.Parse(os.Getenv("RUSTFS_ENDPOINT"))
	if err != nil {
		t.Fatal(err)
	}
	client, err := minio.New(endpoint.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("RUSTFS_ACCESS_KEY"), os.Getenv("RUSTFS_SECRET_KEY"), ""),
		Secure: endpoint.Scheme == "https",
	})
	if err != nil {
		t.Fatal(err)
	}
	key := fmt.Sprintf("%s/%s/%s/%s.json", adapter.Name(), app, version, scanID)
	if err := client.RemoveObject(ctx, store.BucketSBOMs, key, minio.RemoveObjectOptions{}); err != nil {
		t.Fatal(err)
	}
	readBack, err := config.IPFSSBOM.GetSBOM(ctx, key)
	if err != nil || !bytes.Equal(readBack, files[1]) {
		t.Fatalf("IPFS did not return the original SBOM: %v", err)
	}
	request := model.VerifyRequest{Network: adapter.Name(), App: app, Version: version, RecordID: attested.RecordID}
	verifyStatus := func(want string) {
		t.Helper()
		response := post("/api/v1/verify", request, http.StatusOK)
		var result model.VerifyResponse
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil || result.Status != want {
			t.Fatalf("wanted %s, got %q: %v", want, result.Status, err)
		}
	}
	verifyStatus(model.StatusValid)
	unavailable := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer unavailable.Close()
	config.IPFSSBOM, err = store.NewIPFSSBOMStore(blob, index, unavailable.URL, app)
	if err != nil {
		t.Fatal(err)
	}
	server = NewServer(config)
	verifyStatus(model.StatusFileUnavailable)
	attestation.ScanID += "-no-pin"
	response = post("/api/v1/attestations", attestation, http.StatusCreated, http.StatusAccepted)
	if err := json.Unmarshal(response.Body.Bytes(), &attested); err != nil || attested.TxRef == "" {
		t.Fatalf("unavailable IPFS blocked the write: %+v, %v", attested, err)
	}
	t.Log("verified IPFS fallback, unavailable sources, and non-blocking pin failure")
}
