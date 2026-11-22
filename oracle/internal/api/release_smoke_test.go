package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/brii26/eth-iota-attestation/oracle/internal/chain"
	"github.com/brii26/eth-iota-attestation/oracle/internal/model"
	"github.com/brii26/eth-iota-attestation/oracle/internal/store"
	"github.com/brii26/eth-iota-attestation/oracle/internal/verify"
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
	registry := os.Getenv("ETHEREUM_PRIVATE_CONTRACT_ADDRESS")
	adapter, err := chain.NewEthereumAdapter(ctx, chain.EthereumConfig{
		Name: "ethereum-private", Endpoints: []string{os.Getenv("ETHEREUM_PRIVATE_RPC_URL")},
		ContractAddress: registry, PrivateKeyHex: os.Getenv("ETHEREUM_ORACLE_PRIVATE_KEY"), Index: index,
	})
	if err != nil {
		t.Fatal(err)
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
	server := NewServer(Config{S3Store: blob, ChainAdapter: adapter, Index: index, VerdaccioClient: source,
		AuthorizePipeline: func(_ context.Context, candidate, name string) bool {
			return candidate == token && name == app
		}})
	post := func(path string, data any, want int) *httptest.ResponseRecorder {
		t.Helper()
		body, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}
		request := httptest.NewRequestWithContext(ctx, http.MethodPost, path, bytes.NewReader(body))
		request.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != want {
			t.Fatalf("%s got HTTP %d, wanted %d: %s", path, response.Code, want, response.Body.String())
		}
		return response
	}
	response := post("/api/v1/attestations", model.AttestationRequest{
		Network: adapter.Name(), App: app, Version: version, ScanID: scanID,
		ClaimedHashes: model.ClaimedHashes{Artifact: hashes[0], SBOM: hashes[1], Scan: hashes[2]},
		SBOMBytes:     files[1], ScanBytes: files[2],
	}, http.StatusCreated)
	var attested model.AttestationStatus
	if err := json.Unmarshal(response.Body.Bytes(), &attested); err != nil || attested.TxRef == "" || attested.RecordID == "" {
		t.Fatalf("missing transaction receipt: %+v, %v", attested, err)
	}
	response = post("/api/v1/verify", model.VerifyRequest{
		Network: adapter.Name(), App: app, Version: version, RecordID: attested.RecordID,
	}, http.StatusOK)
	var verified model.VerifyResponse
	if err := json.Unmarshal(response.Body.Bytes(), &verified); err != nil || verified.Status != model.StatusValid ||
		verified.Stale || verified.GrypeDBBuiltAt == nil || verified.ScanID != scanID {
		t.Fatalf("release verification failed: %+v, %v", verified, err)
	}
	t.Logf("verified %s@%s in transaction %s", app, version, attested.TxRef)
}
