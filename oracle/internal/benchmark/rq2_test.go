package benchmark

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/brii26/eth-iota-attestation/oracle/internal/api"
	"github.com/brii26/eth-iota-attestation/oracle/internal/chain"
	"github.com/brii26/eth-iota-attestation/oracle/internal/model"
	"github.com/brii26/eth-iota-attestation/oracle/internal/store"
	"github.com/brii26/eth-iota-attestation/oracle/internal/verify"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type rq2Config struct {
	Manifest    string `json:"manifest"`
	Output      string `json:"output"`
	Index       string `json:"index"`
	Bursts      []int  `json:"bursts"`
	Repetitions int    `json:"repetitions"`
}

type rq2Sample struct {
	Network             string             `json:"network"`
	SizeClass           string             `json:"sizeClass"`
	App                 string             `json:"app"`
	Version             string             `json:"version"`
	ScanID              string             `json:"scanId"`
	TxRef               string             `json:"txRef,omitempty"`
	RecordID            string             `json:"recordId,omitempty"`
	Failure             string             `json:"failure,omitempty"`
	FinalityError       string             `json:"finalityError,omitempty"`
	SBOMBytes           int64              `json:"sbomBytes"`
	TxBytes             uint64             `json:"transactionBytes,omitempty"`
	PayloadBytes        int                `json:"payloadBytes,omitempty"`
	GasUsed             uint64             `json:"gasUsed,omitempty"`
	TotalFeeWei         string             `json:"totalFeeWei,omitempty"`
	IOTA                *model.IOTAMetrics `json:"iota,omitempty"`
	SubmittedAt         time.Time          `json:"submittedAt"`
	IncludedAt          *time.Time         `json:"includedAt,omitempty"`
	FinalizedAt         *time.Time         `json:"finalizedAt,omitempty"`
	OracleBeforeChainMS float64            `json:"oracleBeforeChainMS,omitempty"`
	HashingMS           float64            `json:"inputReadAndHashMS"`
	ReadMS              float64            `json:"readMS,omitempty"`
	VerifyMS            float64            `json:"verifyMS,omitempty"`
}

type rq2SampleContext struct{}
type timedAdapter struct{ chain.Adapter }

func (a timedAdapter) WriteRecord(ctx context.Context, record chain.Record) (*chain.Receipt, error) {
	if sample, ok := ctx.Value(rq2SampleContext{}).(*rq2Sample); ok {
		sample.OracleBeforeChainMS = milliseconds(time.Since(sample.SubmittedAt))
	}
	return a.Adapter.WriteRecord(ctx, record)
}

type rq2Batch struct {
	Network    string       `json:"network"`
	SizeClass  string       `json:"sizeClass"`
	Burst      int          `json:"burst"`
	Round      int          `json:"round"`
	Throughput float64      `json:"throughputPerSecond"`
	Samples    []*rq2Sample `json:"samples"`
}

type rq2Stats struct {
	Trials   int     `json:"trials"`
	Failures int     `json:"failures"`
	Min      float64 `json:"minMS"`
	Median   float64 `json:"medianMS"`
	P95      float64 `json:"p95MS"`
	Max      float64 `json:"maxMS"`
}

func TestRQ2TestnetComparison(t *testing.T) {
	configPath := os.Getenv("ATTESTATION_RQ2_CONFIG")
	if configPath == "" {
		t.Skip("set ATTESTATION_RQ2_CONFIG to run the funded testnet benchmark")
	}
	var cfg rq2Config
	readJSON(t, configPath, &cfg)
	if cfg.Manifest == "" || cfg.Output == "" || cfg.Index == "" || cfg.Repetitions < 1 ||
		len(cfg.Bursts) == 0 || !slices.Contains(cfg.Bursts, 1) {
		t.Fatal("manifest, output, index, repetitions and burst sizes including 1 are required")
	}
	for _, count := range cfg.Bursts {
		if count < 1 || count > 100 {
			t.Fatal("burst sizes must be 1 to 100")
		}
	}
	var dataset Dataset
	readJSON(t, cfg.Manifest, &dataset)
	if dataset.SchemaVersion != 1 || dataset.Registry == "" {
		t.Fatal("use a published version 1 dataset manifest")
	}
	classes := map[string][]Fixture{}
	for _, fixture := range dataset.Fixtures {
		classes[fixture.SizeClass] = append(classes[fixture.SizeClass], fixture)
	}
	for _, sizeClass := range []string{"small", "medium", "large"} {
		if len(classes[sizeClass]) == 0 {
			t.Fatalf("dataset has no %s fixtures", sizeClass)
		}
	}
	started := time.Now().UTC()
	idx, err := store.NewSQLiteIndex(cfg.Index)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()

	rpcURL := os.Getenv("ETHEREUM_SEPOLIA_RPC_URL")
	endpoints := []string{rpcURL}
	if backup := os.Getenv("ETHEREUM_SEPOLIA_BACKUP_RPC_URL"); backup != "" {
		endpoints = append(endpoints, backup)
	}
	ethereum, err := chain.NewEthereumAdapter(t.Context(), chain.EthereumConfig{
		Name: "ethereum-sepolia", Endpoints: endpoints, Index: idx,
		ContractAddress: os.Getenv("ETHEREUM_SEPOLIA_CONTRACT_ADDRESS"),
		PrivateKeyHex:   os.Getenv("ETHEREUM_ORACLE_PRIVATE_KEY"),
	})
	if err != nil {
		t.Fatal(err)
	}
	ethRPC, err := ethclient.DialContext(t.Context(), rpcURL)
	if err != nil {
		t.Fatal(err)
	}
	defer ethRPC.Close()
	iota, err := chain.NewIOTAAdapter(t.Context(), os.Getenv("IOTA_SERVICE_URL"), os.Getenv("IOTA_SERVICE_TOKEN"),
		strings.Split(os.Getenv("IOTA_TRUSTED_SIGNERS"), ","), idx)
	if err != nil {
		t.Fatal(err)
	}
	if iota.Name() != "iota-testnet" {
		t.Fatal("the comparison requires IOTA testnet")
	}
	blob, err := store.NewS3Store(t.Context(), store.S3Config{
		Endpoint: os.Getenv("RUSTFS_ENDPOINT"), AccessKey: os.Getenv("RUSTFS_ACCESS_KEY"),
		SecretKey: os.Getenv("RUSTFS_SECRET_KEY"),
	})
	if err != nil {
		t.Fatal(err)
	}
	allowed := make(map[string]bool)
	for _, item := range dataset.Fixtures {
		allowed[item.App] = true
	}

	router := chain.NewRouter()
	router.Register(timedAdapter{ethereum})
	router.Register(timedAdapter{iota})
	source := verify.NewVerdaccioClient(dataset.Registry)
	token := fmt.Sprintf("%x", randomID())
	server := api.NewServer(api.Config{S3Store: blob, Index: idx, Router: router, VerdaccioClient: source,
		AuthorizePipeline: func(_ context.Context, candidate, app string) bool {
			return candidate == token && allowed[app]
		}})
	ctx, cancel := context.WithCancel(t.Context())
	var maintenance sync.WaitGroup
	maintenance.Add(1)
	defer maintenance.Wait()
	defer cancel()
	go func() {
		defer maintenance.Done()
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			recover, end := context.WithTimeout(ctx, 20*time.Second)
			_ = ethereum.Reconcile(recover)
			end()
			recover, end = context.WithTimeout(ctx, 20*time.Second)
			_ = iota.Reconcile(recover)
			end()
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()

	var results []rq2Batch
	var finalized sync.WaitGroup
	networks := []chain.Adapter{ethereum, iota}
	for round := 0; round < cfg.Repetitions; round++ {
		for _, size := range []string{"small", "medium", "large"} {
			for _, burst := range cfg.Bursts {
				for offset := range networks {
					target := networks[(round+offset)%len(networks)]
					batch := rq2Batch{Network: target.Name(), SizeClass: size, Burst: burst, Round: round + 1, Samples: make([]*rq2Sample, burst)}
					var writes sync.WaitGroup
					begin := time.Now()
					for index := range batch.Samples {
						fixture := classes[size][(round+index)%len(classes[size])]
						writes.Add(1)
						go func() {
							defer writes.Done()
							batch.Samples[index] = submitRQ2(ctx, server, target, token, filepath.Dir(cfg.Manifest), fixture)
						}()
					}
					writes.Wait()
					included := 0
					for _, sample := range batch.Samples {
						if sample.IncludedAt == nil {
							continue
						}
						included++
						finalized.Add(1)
						go func() {
							defer finalized.Done()
							followFinality(ctx, ethRPC, iota, sample)
						}()
					}
					batch.Throughput = float64(included) / time.Since(begin).Seconds()
					results = append(results, batch)
				}
			}
		}
	}
	finalized.Wait()
	statistics := summarizeRQ2(results)
	hostname, _ := os.Hostname()
	report := map[string]any{
		"schemaVersion": 1, "startedAt": started, "finishedAt": time.Now().UTC(),
		"dataset": cfg.Manifest, "datasetVersion": dataset.Version, "results": results,
		"statistics": statistics, "client": map[string]string{"host": hostname, "os": runtime.GOOS, "arch": runtime.GOARCH, "go": runtime.Version()},
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfg.Output, append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote %d workload batches to %s", len(results), cfg.Output)
	for _, batch := range results {
		for _, sample := range batch.Samples {
			if sample.Failure != "" || sample.FinalityError != "" {
				t.Errorf("%s %s: %s %s", sample.Network, sample.ScanID, sample.Failure, sample.FinalityError)
			}
		}
	}
}

func submitRQ2(parent context.Context, server http.Handler, network chain.Adapter, token, directory string, fixture Fixture) *rq2Sample {
	sample := &rq2Sample{Network: network.Name(), App: fixture.App, Version: fixture.Version,
		SizeClass: fixture.SizeClass, SBOMBytes: fixture.SBOM.Bytes, ScanID: fmt.Sprintf("rq2-%x", randomID())}
	fail := func(err error) *rq2Sample { sample.Failure = err.Error(); return sample }
	start := time.Now()
	_, err := checkedFile(directory, fixture.Artifact)
	if err != nil {
		return fail(err)
	}
	sbom, err := checkedFile(directory, fixture.SBOM)
	if err != nil {
		return fail(err)
	}
	scan, err := checkedFile(directory, fixture.Scan)
	if err != nil {
		return fail(err)
	}
	sample.HashingMS = milliseconds(time.Since(start))
	ctx, cancel := context.WithTimeout(parent, 4*time.Minute)
	defer cancel()
	request := model.AttestationRequest{Network: sample.Network, App: fixture.App, Version: fixture.Version,
		ScanID: sample.ScanID, SBOMBytes: sbom, ScanBytes: scan,
		ClaimedHashes: model.ClaimedHashes{Artifact: fixture.Artifact.SHA256, SBOM: fixture.SBOM.SHA256, Scan: fixture.Scan.SHA256}}
	sample.SubmittedAt = time.Now().UTC()
	ctx = context.WithValue(ctx, rq2SampleContext{}, sample)
	var submission model.AttestationStatus
	if err := localJSON(ctx, server, http.MethodPost, "/api/v1/attestations", token, request, &submission); err != nil {
		return fail(err)
	}
	sample.TxRef = submission.TxRef
	for submission.Status == "pending" {
		if err := pause(ctx, time.Second); err != nil {
			return fail(err)
		}
		if err := localJSON(ctx, server, http.MethodGet, "/api/v1/attestations/"+sample.Network+"/"+sample.ScanID, "", nil, &submission); err != nil {
			return fail(err)
		}
	}
	sample.TxRef, sample.RecordID = submission.TxRef, submission.RecordID
	if submission.Status != "included" {
		return fail(fmt.Errorf("submission status %s", submission.Status))
	}
	now := time.Now().UTC()
	sample.IncludedAt = &now
	query := chain.Query{App: fixture.App, Version: fixture.Version}
	if sample.Network == "iota-testnet" {
		query.ExternalRecordID = sample.RecordID
	} else {
		query.RecordID, _ = new(big.Int).SetString(sample.RecordID, 10)
		if query.RecordID == nil {
			return fail(errors.New("invalid Ethereum record ID"))
		}
	}
	start = time.Now()
	if _, err := network.ReadRecord(ctx, query); err != nil {
		return fail(err)
	}
	sample.ReadMS = milliseconds(time.Since(start))
	start = time.Now()
	var result model.VerifyResponse
	err = localJSON(ctx, server, http.MethodPost, "/api/v1/verify", "", model.VerifyRequest{
		Network: sample.Network, App: fixture.App, Version: fixture.Version, RecordID: sample.RecordID,
	}, &result)
	sample.VerifyMS = milliseconds(time.Since(start))
	if err != nil {
		return fail(err)
	}
	if result.Status != model.StatusValid {
		return fail(fmt.Errorf("verification status %s", result.Status))
	}
	return sample
}

func followFinality(parent context.Context, eth *ethclient.Client, iota *chain.IOTAAdapter, sample *rq2Sample) {
	ctx, cancel := context.WithTimeout(parent, 32*time.Minute)
	defer cancel()
	hash := common.HexToHash(sample.TxRef)
	var failure error
	for ctx.Err() == nil {
		attempt, stop := context.WithTimeout(ctx, 8*time.Second)
		if sample.Network == "ethereum-sepolia" {
			receipt, err := eth.TransactionReceipt(attempt, hash)
			failure = err
			if err == nil {
				sample.GasUsed = receipt.GasUsed
				if receipt.EffectiveGasPrice != nil {
					sample.TotalFeeWei = new(big.Int).Mul(receipt.EffectiveGasPrice, new(big.Int).SetUint64(receipt.GasUsed)).String()
				}
				if tx, _, err := eth.TransactionByHash(attempt, hash); err == nil {
					sample.TxBytes, sample.PayloadBytes = uint64(tx.Size()), len(tx.Data())
				}
				head, err := eth.HeaderByNumber(attempt, big.NewInt(int64(rpc.FinalizedBlockNumber)))
				failure = err
				if err == nil && head.Number.Cmp(receipt.BlockNumber) >= 0 {
					if block, err := eth.HeaderByNumber(attempt, receipt.BlockNumber); err == nil && block.Hash() == receipt.BlockHash {
						now := time.Now().UTC()
						sample.FinalizedAt = &now
					}
				}
			}
		} else {
			receipt, err := iota.ReadReceipt(attempt, sample.RecordID)
			failure = err
			if err == nil {
				sample.TxBytes = receipt.TxSize
				sample.IOTA = &model.IOTAMetrics{Checkpoint: receipt.Checkpoint, Computation: receipt.GasComputation,
					Storage: receipt.GasStorage, Rebate: receipt.GasRebate, FinalizedAt: receipt.FinalizedAt,
					TransactionSize: receipt.TxSize, DeleteUnlockAt: receipt.DeleteUnlockAt}
				sample.PayloadBytes = receipt.RecordBytes
				if receipt.FinalizedAt != nil {
					now := time.Now().UTC()
					sample.FinalizedAt = &now
				}
			}
		}
		stop()
		if sample.FinalizedAt != nil {
			return
		}
		_ = pause(ctx, 8*time.Second)
	}
	sample.FinalityError = errors.Join(ctx.Err(), failure).Error()
}

func summarizeRQ2(batches []rq2Batch) map[string]map[string]rq2Stats {
	groups := make(map[string][]*rq2Sample)
	for _, batch := range batches {
		label := fmt.Sprintf("%s/%s/%d", batch.Network, batch.SizeClass, batch.Burst)
		groups[label] = append(groups[label], batch.Samples...)
	}
	result := make(map[string]map[string]rq2Stats)
	for label, samples := range groups {
		metrics := map[string]func(*rq2Sample) float64{
			"inclusion": func(s *rq2Sample) float64 {
				if s.IncludedAt == nil {
					return 0
				}
				return milliseconds(s.IncludedAt.Sub(s.SubmittedAt))
			},
			"finality": func(s *rq2Sample) float64 {
				if s.FinalizedAt == nil {
					return 0
				}
				return milliseconds(s.FinalizedAt.Sub(s.SubmittedAt))
			},
			"inputReadAndHash":  func(s *rq2Sample) float64 { return s.HashingMS },
			"oracleBeforeChain": func(s *rq2Sample) float64 { return s.OracleBeforeChainMS },
			"read":              func(s *rq2Sample) float64 { return s.ReadMS },
			"verify":            func(s *rq2Sample) float64 { return s.VerifyMS },
		}
		result[label] = make(map[string]rq2Stats)
		for name, value := range metrics {
			stats := rq2Stats{Trials: len(samples)}
			var values []float64
			for _, sample := range samples {
				if measured := value(sample); measured > 0 {
					values = append(values, measured)
				} else {
					stats.Failures++
				}
			}
			if len(values) > 0 {
				slices.Sort(values)
				stats.Min, stats.Max = values[0], values[len(values)-1]
				stats.Median = (values[(len(values)-1)/2] + values[len(values)/2]) / 2
				stats.P95 = values[int(math.Ceil(.95*float64(len(values))))-1]
			}
			result[label][name] = stats
		}
	}
	return result
}

func checkedFile(directory string, record File) ([]byte, error) {
	if !filepath.IsLocal(record.Path) {
		return nil, errors.New("dataset file path must be local")
	}
	content, err := os.ReadFile(filepath.Join(directory, record.Path))
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(content)
	if int64(len(content)) != record.Bytes || hex.EncodeToString(hash[:]) != record.SHA256 {
		return nil, fmt.Errorf("dataset fingerprint mismatch: %s", record.Path)
	}
	return content, nil
}

func localJSON(ctx context.Context, server http.Handler, method, path, token string, body, result any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	request := httptest.NewRequestWithContext(ctx, method, path, bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code < 200 || response.Code >= 300 {
		return fmt.Errorf("HTTP %d: %s", response.Code, response.Body.String())
	}
	return json.Unmarshal(response.Body.Bytes(), result)
}

func pause(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func milliseconds(value time.Duration) float64 { return float64(value) / float64(time.Millisecond) }

func readJSON(t *testing.T, path string, result any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, result); err != nil {
		t.Fatal(err)
	}
}

func randomID() [16]byte {
	var result [16]byte
	rand.Read(result[:])
	return result
}

func TestRQ2StatisticsKeepFailedTrials(t *testing.T) {
	started := time.Unix(0, 0)
	mid := started.Add(200 * time.Millisecond)
	high := started.Add(300 * time.Millisecond)
	low := started.Add(100 * time.Millisecond)
	stats := summarizeRQ2([]rq2Batch{{Network: "ethereum-sepolia", SizeClass: "small", Burst: 4,
		Samples: []*rq2Sample{
			{SubmittedAt: started, IncludedAt: &high},
			{SubmittedAt: started, IncludedAt: &low},
			{SubmittedAt: started, IncludedAt: &mid},
			{SubmittedAt: started, Failure: "unavailable"},
		}}})
	got := stats["ethereum-sepolia/small/4"]["inclusion"]
	want := rq2Stats{Trials: 4, Failures: 1, Min: 100, Median: 200, P95: 300, Max: 300}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}
