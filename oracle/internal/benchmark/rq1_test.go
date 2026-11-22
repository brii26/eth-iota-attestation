package benchmark

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/brii26/eth-iota-attestation/oracle/internal/api"
	"github.com/brii26/eth-iota-attestation/oracle/internal/chain"
	"github.com/brii26/eth-iota-attestation/oracle/internal/model"
	"github.com/brii26/eth-iota-attestation/oracle/internal/store"
	"github.com/brii26/eth-iota-attestation/oracle/internal/verify"
	"github.com/brii26/eth-iota-attestation/smart-contract/registry"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type rq1Case struct {
	Network   string `json:"network"`
	App       string `json:"app"`
	Version   string `json:"version"`
	Target    string `json:"target"`
	Scenario  string `json:"scenario"`
	Component string `json:"component"`
	Expected  string `json:"expected"`
	Observed  string `json:"observed"`
}

type alteredBlobs struct {
	verify.BlobStore
	file    string
	variant int
}

func alter(data []byte, err error, flip bool, variant int) ([]byte, error) {
	if err != nil || !flip {
		return data, err
	}
	altered := bytes.Clone(data)
	if len(altered) == 0 {
		return []byte{'\n'}, nil
	}
	altered[variant%len(altered)] ^= 1
	return altered, nil
}

func (s alteredBlobs) GetSBOM(ctx context.Context, key string) ([]byte, error) {
	data, err := s.BlobStore.GetSBOM(ctx, key)
	return alter(data, err, s.file == "sbom", s.variant)
}

func (s alteredBlobs) GetScan(ctx context.Context, key string) ([]byte, error) {
	data, err := s.BlobStore.GetScan(ctx, key)
	return alter(data, err, s.file == "scan", s.variant)
}

type alteredPackage struct {
	verify.ArtifactFetcher
	changed bool
	variant int
}

func (s alteredPackage) FetchArtifact(ctx context.Context, app, version string) ([]byte, error) {
	data, err := s.ArtifactFetcher.FetchArtifact(ctx, app, version)
	return alter(data, err, s.changed, s.variant)
}

const staleGrypeScan = `{"descriptor":{"db":{"built":"2000-01-01T00:00:00Z"}}}`

func TestRQ1IntegrityMatrix(t *testing.T) {
	configPath := os.Getenv("ATTESTATION_RQ1_CONFIG")
	if configPath == "" {
		t.Skip("set ATTESTATION_RQ1_CONFIG to run the ledger integrity matrix")
	}
	var cfg struct {
		Manifest string `json:"manifest"`
		Output   string `json:"output"`
	}
	readJSON(t, configPath, &cfg)
	if cfg.Manifest == "" || cfg.Output == "" {
		t.Fatal("manifest and output are required")
	}
	var dataset Dataset
	readJSON(t, cfg.Manifest, &dataset)
	if dataset.SchemaVersion != 1 || dataset.Registry == "" || len(dataset.Fixtures) < 10 {
		t.Fatal("use a published dataset with at least 10 fixtures")
	}
	indexPath := filepath.Join(t.TempDir(), "rq1.sqlite")
	index, err := store.NewSQLiteIndex(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()
	temporaryIndex, err := sql.Open("sqlite", indexPath)
	if err != nil {
		t.Fatal(err)
	}
	defer temporaryIndex.Close()
	ctx := t.Context()
	rpc := os.Getenv("ETHEREUM_PRIVATE_RPC_URL")
	ethKey := os.Getenv("ETHEREUM_ORACLE_PRIVATE_KEY")
	eth, err := chain.NewEthereumAdapter(ctx, chain.EthereumConfig{
		Name: "ethereum-private", Endpoints: []string{rpc}, PrivateKeyHex: ethKey,
		ContractAddress: os.Getenv("ETHEREUM_PRIVATE_CONTRACT_ADDRESS"), Index: index,
	})
	if err != nil {
		t.Fatal(err)
	}
	directETH, err := chain.NewEthereumAdapter(ctx, chain.EthereumConfig{
		Name: eth.Name(), Endpoints: []string{rpc}, PrivateKeyHex: ethKey,
		ContractAddress: os.Getenv("ETHEREUM_PRIVATE_CONTRACT_ADDRESS"),
	})
	if err != nil {
		t.Fatal(err)
	}

	iota, err := chain.NewIOTAAdapter(ctx, os.Getenv("IOTA_SERVICE_URL"), os.Getenv("IOTA_SERVICE_TOKEN"),
		strings.Split(os.Getenv("IOTA_TRUSTED_SIGNERS"), ","), index)
	if err != nil {
		t.Fatal(err)
	}
	revokedIOTA, err := chain.NewIOTAAdapter(ctx, os.Getenv("IOTA_SERVICE_URL"), os.Getenv("IOTA_SERVICE_TOKEN"),
		[]string{"0x" + strings.Repeat("0", 64)}, index)
	if err != nil {
		t.Fatal(err)
	}
	outside, err := store.NewSQLiteIndex(filepath.Join(t.TempDir(), "outside-application.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer outside.Close()
	directIOTA, err := chain.NewIOTAAdapter(ctx, os.Getenv("IOTA_SERVICE_URL"), os.Getenv("IOTA_SERVICE_TOKEN"),
		strings.Split(os.Getenv("IOTA_TRUSTED_SIGNERS"), ","), outside)
	if err != nil {
		t.Fatal(err)
	}
	blobs, err := store.NewS3Store(ctx, store.S3Config{
		Endpoint: os.Getenv("RUSTFS_ENDPOINT"), AccessKey: os.Getenv("RUSTFS_ACCESS_KEY"),
		SecretKey: os.Getenv("RUSTFS_SECRET_KEY"),
	})
	if err != nil {
		t.Fatal(err)
	}
	client, err := ethclient.DialContext(ctx, rpc)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	address := os.Getenv("BASELINE_ETHEREUM_CONTRACT")
	if !common.IsHexAddress(address) {
		t.Fatal("configure the experimental Ethereum baseline contract")
	}
	baseline, err := registry.NewBaselineRegistry(common.HexToAddress(address), client)
	if err != nil {
		t.Fatal(err)
	}
	key, err := crypto.HexToECDSA(strings.TrimPrefix(ethKey, "0x"))
	if err != nil {
		t.Fatal(err)
	}
	writer, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(1337))
	if err != nil {
		t.Fatal(err)
	}
	adminKey, err := crypto.HexToECDSA(strings.TrimPrefix(os.Getenv("ETHEREUM_ADMIN_PRIVATE_KEY"), "0x"))
	if err != nil {
		t.Fatal("set a private Geth administrator key for the dedicated benchmark registry")
	}
	admin, err := bind.NewKeyedTransactorWithChainID(adminKey, big.NewInt(1337))
	if err != nil {
		t.Fatal(err)
	}
	fullRegistry, err := registry.NewAttestationRegistry(eth.ContractAddress(), client)
	if err != nil {
		t.Fatal(err)
	}
	owner, err := fullRegistry.Admin(&bind.CallOpts{Context: ctx})
	if err != nil || owner != admin.From {
		t.Fatalf("benchmark account is not the registry administrator: %v", err)
	}
	router := chain.NewRouter()
	router.Register(eth)
	router.Register(iota)
	source := verify.NewVerdaccioClient(dataset.Registry)
	token := fmt.Sprintf("%x", randomID())
	server := api.NewServer(api.Config{S3Store: blobs, Index: index, Router: router, VerdaccioClient: source,
		AuthorizePipeline: func(_ context.Context, candidate, _ string) bool { return candidate == token }})
	start := time.Now().UTC()
	var results []rq1Case
	type baselineCase struct {
		network chain.Adapter
		fixture Fixture
		variant int
	}
	var baselines []baselineCase
	for variant, fixture := range dataset.Fixtures {
		for _, network := range []chain.Adapter{eth, iota} {
			add := func(scenario, component, want, observed string) {
				results = append(results, rq1Case{network.Name(), fixture.App, fixture.Version, "attestation", scenario, component, want, observed})
			}
			t.Logf("fixture %d/%d: %s %s", variant+1, len(dataset.Fixtures), network.Name(), fixture.App)
			baselines = append(baselines, baselineCase{network: network, fixture: fixture, variant: variant})
			sample := submitRQ2(ctx, server, network, token, filepath.Dir(cfg.Manifest), fixture)
			if sample.Failure != "" {
				add("valid", "oracle", model.StatusValid, sample.Failure)
				continue
			}
			add("valid", "oracle", model.StatusValid, model.StatusValid)
			request := model.VerifyRequest{Network: network.Name(), App: fixture.App, Version: fixture.Version, RecordID: sample.RecordID}
			for _, item := range []struct{ file, status string }{
				{"artifact", model.StatusArtifactMismatch}, {"sbom", model.StatusSBOMMismatch}, {"scan", model.StatusScanMismatch},
			} {
				engine := verify.NewEngine(verify.EngineConfig{
					ArtifactFetcher: alteredPackage{source, item.file == "artifact", variant}, BlobStore: alteredBlobs{blobs, item.file, variant},
				})
				add(item.file+"-changed", "verifier", item.status, integrityResult(ctx, engine, network, request))
			}
			engine := verify.NewEngine(verify.EngineConfig{ArtifactFetcher: source, BlobStore: blobs})
			otherVersion := request
			otherVersion.Version += "-different"
			add("wrong-version", "record-binding", model.StatusRecordNotFound, integrityResult(ctx, engine, network, otherVersion))
			if network.Name() == eth.Name() {
				add("direct-replay", "ethereum-contract", "ScanIdAlreadyUsed", ethereumDirectWrite(ctx, directETH, fixture, sample.ScanID, "ScanIdAlreadyUsed"))
				observed, write, err := verifyAfterRevocation(ctx, client, fullRegistry, admin, writer.From, engine, eth, request, func(ctx context.Context) string {
					scanID := fmt.Sprintf("rq1-untrusted-%x", randomID())
					return ethereumDirectWrite(ctx, directETH, fixture, scanID, "UnauthorizedOracle")
				})
				if err != nil {
					t.Fatal(err)
				}
				add("untrusted-signer", "ethereum-contract", model.StatusUntrustedSigner, observed)
				add("unauthorized-oracle-write", "ethereum-contract", "UnauthorizedOracle", write)
			} else {
				add("direct-replay", "iota-ledger-with-separate-index", "accepted",
					iotaReplay(ctx, directIOTA, outside, fixture, sample.ScanID))
				add("untrusted-signer", "iota-policy", model.StatusUntrustedSigner, integrityResult(ctx, engine, revokedIOTA, request))
			}

			sbom, err := checkedFile(filepath.Dir(cfg.Manifest), fixture.SBOM)
			if err != nil {
				t.Fatal(err)
			}
			scan, err := checkedFile(filepath.Dir(cfg.Manifest), fixture.Scan)
			if err != nil {
				t.Fatal(err)
			}
			payload := model.AttestationRequest{Network: network.Name(), App: fixture.App, Version: fixture.Version, ScanID: sample.ScanID,
				SBOMBytes: append(bytes.Clone(sbom), '\n'), ScanBytes: scan,
				ClaimedHashes: model.ClaimedHashes{Artifact: fixture.Artifact.SHA256, Scan: fixture.Scan.SHA256}}
			_, payload.ClaimedHashes.SBOM = verify.ComputeSHA256(payload.SBOMBytes)
			add("replay-scan-id", "oracle-index", model.ErrCodeDuplicateScanID, submissionCode(ctx, server, token, payload))
			payload.ScanID = fmt.Sprintf("rq1-%x", randomID())
			add("pipeline-not-authorized", "oracle-auth", model.ErrCodeUnauthorizedPipeline, submissionCode(ctx, server, "denied", payload))
			payload.ScanBytes = []byte(staleGrypeScan)
			_, payload.ClaimedHashes.Scan = verify.ComputeSHA256(payload.ScanBytes)
			add("stale-scan", "oracle-freshness", model.ErrCodeStaleScan, submissionCode(ctx, server, token, payload))

			newer := submitRQ2(ctx, server, network, token, filepath.Dir(cfg.Manifest), fixture)
			wantLatest := "newest-ledger-record"
			if network.Name() == iota.Name() {
				wantLatest = "older-indexed-record"
			}
			if newer.Failure != "" {
				add("stale-index", "record-discovery", wantLatest, newer.Failure)
			} else {
				add("stale-index", "record-discovery", wantLatest,
					readWithOlderIndex(ctx, temporaryIndex, network, fixture, sample.RecordID, newer.RecordID))
			}
		}
	}
	for idx, planned := range baselines {
		fixture, network := planned.fixture, planned.network
		t.Logf("baseline %d/%d: %s %s", idx+1, len(baselines), network.Name(), fixture.App)
		add := func(scenario, component, want, observed string) {
			results = append(results, rq1Case{network.Name(), fixture.App, fixture.Version, "baseline", scenario, component, want, observed})
		}
		hashes := chain.BaselineHashes{
			ArtifactHash: "0x" + fixture.Artifact.SHA256, SBOMHash: "0x" + fixture.SBOM.SHA256, ScanHash: "0x" + fixture.Scan.SHA256,
		}
		record := func(hashes chain.BaselineHashes) baselineRead {
			return writeBaseline(ctx, client, baseline, writer, iota, network, hashes)
		}
		recorded := record(hashes)
		add("valid", "ledger", "found", recorded(ctx, hashes, iota))
		add("wrong-version", "three-hash-membership", "found", recorded(ctx, hashes, iota))
		for _, item := range []struct {
			scenario string
			file     File
			hash     *string
		}{
			{"artifact-changed", fixture.Artifact, &hashes.ArtifactHash},
			{"sbom-changed", fixture.SBOM, &hashes.SBOMHash},
			{"scan-changed", fixture.Scan, &hashes.ScanHash},
		} {
			data, err := checkedFile(filepath.Dir(cfg.Manifest), item.file)
			if err != nil {
				add(item.scenario, "fixture", "missing", err.Error())
				continue
			}
			altered, _ := alter(data, nil, true, planned.variant)
			_, computed := verify.ComputeSHA256(altered)
			expected := *item.hash
			*item.hash = "0x" + computed
			add(item.scenario, "three-hash-membership", "missing", recorded(ctx, hashes, iota))
			*item.hash = expected
		}
		if network.Name() == eth.Name() {
			add("untrusted-signer", "ethereum-contract", "UnauthorizedWriter", baselineUnauthorized(ctx, client, common.HexToAddress(address), hashes))
		} else {
			add("untrusted-signer", "iota-policy", "missing", recorded(ctx, hashes, revokedIOTA))
		}
		replayHashes := hashes
		replayID := randomID()
		replayHashes.ArtifactHash = crypto.Keccak256Hash(replayID[:]).Hex()
		add("replay-scan-id", "three-hash-ledger-write", "found", record(replayHashes)(ctx, replayHashes, iota))
		stale := hashes
		_, staleHash := verify.ComputeSHA256([]byte(staleGrypeScan))
		stale.ScanHash = "0x" + staleHash
		add("stale-scan", "three-hash-ledger-write", "found", record(stale)(ctx, stale, iota))
		add("pipeline-not-authorized", "three-hash-baseline", "not-applicable", "not-applicable")
		add("stale-index", "three-hash-baseline", "not-applicable", "not-applicable")
	}

	data, err := json.MarshalIndent(map[string]any{
		"schemaVersion": 1, "dataset": cfg.Manifest, "startedAt": start, "finishedAt": time.Now().UTC(), "results": results,
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfg.Output, append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, result := range results {
		if result.Expected != result.Observed {
			t.Errorf("%s %s %s %s: got %q, expected %q", result.Network, result.Target, result.App, result.Scenario, result.Observed, result.Expected)
		}
	}
}

func integrityResult(ctx context.Context, engine *verify.Engine, network chain.Adapter, request model.VerifyRequest) string {
	response, err := engine.Verify(ctx, network, request)
	if err != nil {
		return err.Error()
	}
	return response.Status
}

func submissionCode(ctx context.Context, server http.Handler, token string, payload model.AttestationRequest) string {
	body, err := json.Marshal(payload)
	if err != nil {
		return err.Error()
	}
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/attestations", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, req)
	var result model.ErrorResponse
	if err := json.Unmarshal(response.Body.Bytes(), &result); err == nil && result.Error.Code != "" {
		return result.Error.Code
	}
	return fmt.Sprintf("HTTP %d", response.Code)
}

type baselineRead func(context.Context, chain.BaselineHashes, *chain.IOTAAdapter) string

func changeHash(hash string) string {
	replacement := "1"
	if hash[2] == '1' {
		replacement = "2"
	}
	return hash[:2] + replacement + hash[3:]
}

func baselineResult(found bool, err error) string {
	if err != nil {
		return err.Error()
	}
	if found {
		return "found"
	}
	return "missing"
}

func writeBaseline(parent context.Context, client *ethclient.Client, contract *registry.BaselineRegistry,
	writer *bind.TransactOpts, iota *chain.IOTAAdapter, network chain.Adapter, hashes chain.BaselineHashes) baselineRead {
	ctx, cancel := context.WithTimeout(parent, 3*time.Minute)
	defer cancel()
	failed := func(err error) baselineRead {
		return func(context.Context, chain.BaselineHashes, *chain.IOTAAdapter) string { return err.Error() }
	}
	if network.Name() == "ethereum-private" {
		writer.Context = ctx
		tx, err := contract.AddHashes(writer, common.HexToHash(hashes.ArtifactHash), common.HexToHash(hashes.SBOMHash), common.HexToHash(hashes.ScanHash))
		if err != nil {
			return failed(err)
		}
		receipt, err := bind.WaitMined(ctx, client, tx)
		if err != nil {
			return failed(err)
		}
		if receipt.Status != types.ReceiptStatusSuccessful {
			return failed(fmt.Errorf("baseline write %s reverted", tx.Hash()))
		}
		return func(ctx context.Context, expected chain.BaselineHashes, _ *chain.IOTAAdapter) string {
			return baselineResult(contract.HasHashes(&bind.CallOpts{Context: ctx}, common.HexToHash(expected.ArtifactHash), common.HexToHash(expected.SBOMHash), common.HexToHash(expected.ScanHash)))
		}
	}
	receipt, err := iota.WriteBaseline(ctx, hashes)
	if err != nil {
		return failed(err)
	}
	read, stop := context.WithTimeout(ctx, 30*time.Second)
	defer stop()
	for {
		found, err := iota.VerifyBaseline(read, receipt.RecordID, hashes)
		if err == nil && found {
			break
		}
		if pause(read, time.Second) != nil {
			return failed(fmt.Errorf("baseline object %s unavailable: %s", receipt.RecordID, baselineResult(found, err)))
		}
	}
	return func(ctx context.Context, expected chain.BaselineHashes, reader *chain.IOTAAdapter) string {
		return baselineResult(reader.VerifyBaseline(ctx, receipt.RecordID, expected))
	}
}

func baselineUnauthorized(ctx context.Context, client *ethclient.Client, address common.Address, hashes chain.BaselineHashes) string {
	abi, err := registry.BaselineRegistryMetaData.GetAbi()
	if err != nil {
		return err.Error()
	}
	data, err := abi.Pack("addHashes", common.HexToHash(hashes.ArtifactHash), common.HexToHash(hashes.SBOMHash), common.HexToHash(hashes.ScanHash))
	if err != nil {
		return err.Error()
	}
	_, err = client.CallContract(ctx, ethereum.CallMsg{From: common.BytesToAddress([]byte{1}), To: &address, Data: data}, nil)
	if err == nil {
		return "accepted"
	}
	raw, ok := ethclient.RevertErrorData(err)
	rule := abi.Errors["UnauthorizedWriter"]
	if ok && len(raw) >= 4 && bytes.Equal(raw[:4], rule.ID[:4]) {
		return rule.Name
	}
	return err.Error()
}

func ethereumDirectWrite(ctx context.Context, adapter *chain.EthereumAdapter, fixture Fixture, scanID, expectedRule string) string {
	_, err := adapter.WriteRecord(ctx, chain.Record{
		Network: adapter.Name(), App: fixture.App, Version: fixture.Version, ScanID: scanID,
		ArtifactHash: common.HexToHash(changeHash("0x" + fixture.Artifact.SHA256)),
		SBOMHash:     common.HexToHash(fixture.SBOM.SHA256), ScanHash: common.HexToHash(fixture.Scan.SHA256),
	})
	if err == nil {
		return "accepted"
	}
	revert, ok := ethclient.RevertErrorData(err)
	if !ok || len(revert) < 4 {
		return err.Error()
	}
	abi, err := registry.AttestationRegistryMetaData.GetAbi()
	if err != nil {
		return err.Error()
	}
	rule := abi.Errors[expectedRule]
	if bytes.Equal(revert[:4], rule.ID[:4]) {
		return rule.Name
	}
	return "unexpected-revert"
}

func readWithOlderIndex(ctx context.Context, db *sql.DB, network chain.Adapter, fixture Fixture, older, newer string) string {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err.Error()
	}
	var original time.Time
	if err := tx.QueryRowContext(ctx, "SELECT included_at FROM attestations WHERE network = ? AND record_id = ?",
		network.Name(), older).Scan(&original); err != nil {
		_ = tx.Rollback()
		return err.Error()
	}
	if _, err := tx.ExecContext(ctx, "UPDATE attestations SET included_at = ? WHERE network = ? AND record_id = ?",
		time.Now().UTC().Add(time.Hour), network.Name(), older); err != nil {
		_ = tx.Rollback()
		return err.Error()
	}
	if err := tx.Commit(); err != nil {
		return err.Error()
	}
	result, readErr := network.ReadRecord(ctx, chain.Query{App: fixture.App, Version: fixture.Version})
	if _, err := db.ExecContext(ctx, "UPDATE attestations SET included_at = ? WHERE network = ? AND record_id = ?",
		original, network.Name(), older); err != nil {
		return err.Error()
	}
	if readErr != nil {
		return readErr.Error()
	}
	switch result.RecordID {
	case older:
		return "older-indexed-record"
	case newer:
		return "newest-ledger-record"
	default:
		return "unexpected-record"
	}
}

func verifyAfterRevocation(ctx context.Context, client *ethclient.Client, contract *registry.AttestationRegistry,
	admin *bind.TransactOpts, signer common.Address, engine *verify.Engine, network chain.Adapter,
	request model.VerifyRequest, tryWrite func(context.Context) string) (string, string, error) {
	revokeCtx, stop := context.WithTimeout(ctx, time.Minute)
	defer stop()
	admin.Context = revokeCtx
	revoke, err := contract.RevokeOracle(admin, signer)
	if err != nil {
		return "", "", err
	}
	revoked, err := bind.WaitMined(revokeCtx, client, revoke)
	if err != nil || revoked.Status != types.ReceiptStatusSuccessful {
		return "", "", fmt.Errorf("oracle revocation %s was not confirmed: %v", revoke.Hash(), err)
	}
	observed := integrityResult(revokeCtx, engine, network, request)
	write := tryWrite(revokeCtx)
	restoreCtx, done := context.WithTimeout(context.Background(), time.Minute)
	defer done()
	admin.Context = restoreCtx
	restore, err := contract.AuthorizeOracle(admin, signer)
	if err != nil {
		return "", "", fmt.Errorf("restore benchmark oracle permission: %w", err)
	}
	receipt, err := bind.WaitMined(restoreCtx, client, restore)
	if err != nil || receipt.Status != types.ReceiptStatusSuccessful {
		return "", "", fmt.Errorf("restore benchmark oracle permission %s was not confirmed: %v", restore.Hash(), err)
	}
	return observed, write, nil
}

func iotaReplay(ctx context.Context, adapter *chain.IOTAAdapter, index *store.SQLiteIndex, fixture Fixture, scanID string) string {
	record := store.AttestationRecord{Network: adapter.Name(), App: fixture.App, Version: fixture.Version, ScanID: scanID,
		ArtifactHash: strings.Repeat("1", 64), SBOMHash: fixture.SBOM.SHA256, ScanHash: fixture.Scan.SHA256,
		SubmittedAt: time.Now().UTC(), Status: "pending"}
	if err := index.Save(ctx, record); err != nil {
		return err.Error()
	}
	receipt, err := adapter.WriteRecord(ctx, chain.Record{Network: adapter.Name(), App: fixture.App, Version: fixture.Version, ScanID: scanID,
		ArtifactHash: common.HexToHash(record.ArtifactHash),
		SBOMHash:     common.HexToHash(record.SBOMHash), ScanHash: common.HexToHash(record.ScanHash)})
	if err != nil {
		return err.Error()
	}
	wait, cancel := context.WithTimeout(ctx, 40*time.Second)
	defer cancel()
	for wait.Err() == nil {
		finalized, err := adapter.ReadReceipt(wait, receipt.RecordID)
		if err == nil && finalized.FinalizedAt != nil {
			return "accepted"
		}
		_ = pause(wait, time.Second)
	}
	return "checkpoint-unavailable"
}
