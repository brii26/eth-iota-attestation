package benchmark

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
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
	Scenario  string `json:"scenario"`
	Component string `json:"component"`
	Expected  string `json:"expected"`
	Observed  string `json:"observed"`
}

type alteredBlobs struct {
	verify.BlobStore
	file string
}

func alter(data []byte, err error, flip bool) ([]byte, error) {
	if err == nil && flip {
		data = append(bytes.Clone(data), '\n')
	}
	return data, err
}

func (s alteredBlobs) GetSBOM(ctx context.Context, key string) ([]byte, error) {
	data, err := s.BlobStore.GetSBOM(ctx, key)
	return alter(data, err, s.file == "sbom")
}

func (s alteredBlobs) GetScan(ctx context.Context, key string) ([]byte, error) {
	data, err := s.BlobStore.GetScan(ctx, key)
	return alter(data, err, s.file == "scan")
}

type alteredPackage struct {
	verify.ArtifactFetcher
	changed bool
}

func (s alteredPackage) FetchArtifact(ctx context.Context, app, version string) ([]byte, error) {
	data, err := s.ArtifactFetcher.FetchArtifact(ctx, app, version)
	return alter(data, err, s.changed)
}

type revokedSigner struct{ chain.Adapter }

func (revokedSigner) IsSignerAuthorized(context.Context, string) (bool, error) { return false, nil }

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
	router := chain.NewRouter()
	router.Register(eth)
	router.Register(iota)
	source := verify.NewVerdaccioClient(dataset.Registry)
	token := fmt.Sprintf("%x", randomID())
	server := api.NewServer(api.Config{S3Store: blobs, Index: index, Router: router, VerdaccioClient: source,
		AuthorizePipeline: func(_ context.Context, candidate, _ string) bool { return candidate == token }})
	start := time.Now().UTC()
	var results []rq1Case
	for _, fixture := range dataset.Fixtures {
		for _, network := range []chain.Adapter{eth, iota} {
			add := func(scenario, component, want, observed string) {
				results = append(results, rq1Case{network.Name(), fixture.App, fixture.Version, scenario, component, want, observed})
			}
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
					ArtifactFetcher: alteredPackage{source, item.file == "artifact"}, BlobStore: alteredBlobs{blobs, item.file},
				})
				add(item.file+"-changed", "verifier", item.status, integrityResult(ctx, engine, network, request))
			}
			engine := verify.NewEngine(verify.EngineConfig{ArtifactFetcher: source, BlobStore: blobs})
			otherVersion := request
			otherVersion.Version += "-different"
			add("wrong-version", "record-binding", model.StatusRecordNotFound, integrityResult(ctx, engine, network, otherVersion))
			if network.Name() == eth.Name() {
				add("direct-replay", "ethereum-contract", "ScanIdAlreadyUsed", ethereumReplay(ctx, directETH, fixture, sample.ScanID))
			}
			add("untrusted-signer", "verifier-policy", model.StatusUntrustedSigner, integrityResult(ctx, engine, revokedSigner{network}, request))

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
			payload.ScanBytes = []byte(`{"descriptor":{"db":{"built":"2000-01-01T00:00:00Z"}}}`)
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

			fingerprints := chain.BaselineHashes{
				ArtifactHash: "0x" + fixture.Artifact.SHA256, SBOMHash: "0x" + fixture.SBOM.SHA256, ScanHash: "0x" + fixture.Scan.SHA256,
			}
			matches, err := writeAndCompareBaseline(ctx, client, baseline, writer, iota, network, fingerprints)
			if err != nil {
				add("wrong-version-baseline", "three-hash-baseline", "found", err.Error())
			} else {
				add("wrong-version-baseline", "three-hash-baseline", "found", matches)
				add("changed-hash-baseline", "three-hash-baseline", "missing", "missing")
			}
		}
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
			t.Errorf("%s %s %s: got %q, expected %q", result.Network, result.App, result.Scenario, result.Observed, result.Expected)
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

func writeAndCompareBaseline(parent context.Context, client *ethclient.Client, contract *registry.BaselineRegistry,
	writer *bind.TransactOpts, iota *chain.IOTAAdapter, network chain.Adapter, hashes chain.BaselineHashes) (string, error) {
	ctx, cancel := context.WithTimeout(parent, 3*time.Minute)
	defer cancel()
	changed := hashes
	replacement := byte('1')
	if changed.ArtifactHash[2] == replacement {
		replacement = '2'
	}
	changed.ArtifactHash = changed.ArtifactHash[:2] + string(replacement) + changed.ArtifactHash[3:]
	var found, altered bool
	var err error
	if network.Name() == "ethereum-private" {
		values := [3][32]byte{}
		for index, text := range []string{hashes.ArtifactHash, hashes.SBOMHash, hashes.ScanHash} {
			_, err := hex.Decode(values[index][:], []byte(text[2:]))
			if err != nil {
				return "", err
			}
		}
		writer.Context = ctx
		tx, err := contract.AddHashes(writer, values[0], values[1], values[2])
		if err != nil {
			return "", err
		}
		receipt, err := bind.WaitMined(ctx, client, tx)
		if err != nil {
			return "", err
		}
		if receipt.Status != types.ReceiptStatusSuccessful {
			return "", fmt.Errorf("baseline write %s reverted", tx.Hash())
		}
		call := &bind.CallOpts{Context: ctx}
		found, err = contract.HasHashes(call, values[0], values[1], values[2])
		if err != nil {
			return "", err
		}
		altered, err = contract.HasHashes(call, common.HexToHash(changed.ArtifactHash), values[1], values[2])
	} else {
		receipt, err := iota.WriteBaseline(ctx, hashes)
		if err != nil {
			return "", err
		}
		found, err = iota.VerifyBaseline(ctx, receipt.RecordID, hashes)
		if err != nil {
			return "", err
		}
		altered, err = iota.VerifyBaseline(ctx, receipt.RecordID, changed)
		if err != nil {
			return "", err
		}
	}
	if err != nil {
		return "", err
	}
	if altered {
		return "", fmt.Errorf("baseline accepted a mutated hash")
	}
	if !found {
		return "missing", nil
	}
	return "found", nil
}

func ethereumReplay(ctx context.Context, adapter *chain.EthereumAdapter, fixture Fixture, scanID string) string {
	_, err := adapter.WriteRecord(ctx, chain.Record{
		Network: adapter.Name(), App: fixture.App, Version: fixture.Version, ScanID: scanID,
		ArtifactHash: common.HexToHash(fixture.Artifact.SHA256),
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
	rule := abi.Errors["ScanIdAlreadyUsed"]
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
