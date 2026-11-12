package registry_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"

	"github.com/brii26/eth-iota-attestation/smart-contract/registry"
)

const (
	maxApp     = 214
	maxVersion = 128
	maxScanID  = 128
)

var (
	zeroHash    [32]byte
	artifactH   = filled(0x11)
	sbomH       = filled(0x22)
	scanH       = filled(0x33)
	maxUint256  = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	zeroAddress common.Address
)

func filled(b byte) [32]byte {
	var h [32]byte
	for i := range h {
		h[i] = b
	}
	return h
}

// input is the argument list of addRecord.
type input struct {
	app, version, scanID string
	artifact, sbom, scan [32]byte
}

func record(scanID, app, version string) input {
	return input{app: app, version: version, scanID: scanID, artifact: artifactH, sbom: sbomH, scan: scanH}
}

func defaultRecord() input { return record("scan-1", "demo-lib", "1.0.0") }

type actor struct {
	key  *ecdsa.PrivateKey
	addr common.Address
}

func newActor(t *testing.T) actor {
	t.Helper()
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	return actor{key: key, addr: crypto.PubkeyToAddress(key.PublicKey)}
}

type fixture struct {
	t        *testing.T
	backend  *simulated.Backend
	client   simulated.Client
	contract *registry.AttestationRegistry
	abi      *abi.ABI
	chainID  *big.Int
	admin    actor
	oracle   actor
	stranger actor
}

type txFunc func(*bind.TransactOpts) (*types.Transaction, error)

// newFixture deploys a fresh registry on a simulated chain and authorizes the oracle.
func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{t: t, admin: newActor(t), oracle: newActor(t), stranger: newActor(t)}
	funds := new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))
	alloc := types.GenesisAlloc{}
	for _, a := range []actor{f.admin, f.oracle, f.stranger} {
		alloc[a.addr] = types.Account{Balance: funds}
	}
	f.backend = simulated.NewBackend(alloc)
	t.Cleanup(func() { _ = f.backend.Close() })
	f.client = f.backend.Client()
	id, err := f.client.ChainID(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	f.chainID = id
	parsed, err := registry.AttestationRegistryMetaData.GetAbi()
	if err != nil {
		t.Fatal(err)
	}
	f.abi = parsed

	_, deployTx, contract, err := registry.DeployAttestationRegistry(f.opts(f.admin), f.client)
	if err != nil {
		t.Fatalf("deploy: %v", err)
	}
	f.contract = contract
	f.mine(deployTx)
	f.send(f.admin, func(o *bind.TransactOpts) (*types.Transaction, error) {
		return f.contract.AuthorizeOracle(o, f.oracle.addr)
	})
	return f
}

func (f *fixture) opts(a actor) *bind.TransactOpts {
	f.t.Helper()
	opts, err := bind.NewKeyedTransactorWithChainID(a.key, f.chainID)
	if err != nil {
		f.t.Fatal(err)
	}
	return opts
}

// mine commits a block and returns the receipt of tx, which must have succeeded.
func (f *fixture) mine(tx *types.Transaction) *types.Receipt {
	f.t.Helper()
	f.backend.Commit()
	receipt, err := f.client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		f.t.Fatalf("receipt: %v", err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		f.t.Fatalf("transaction %s failed", tx.Hash())
	}
	return receipt
}

func (f *fixture) send(a actor, fn txFunc) *types.Receipt {
	f.t.Helper()
	tx, err := fn(f.opts(a))
	if err != nil {
		f.t.Fatalf("send: %v", err)
	}
	return f.mine(tx)
}

func (f *fixture) add(a actor, in input) *types.Receipt {
	f.t.Helper()
	return f.send(a, func(o *bind.TransactOpts) (*types.Transaction, error) {
		return f.contract.AddRecord(o, in.app, in.version, in.scanID, in.artifact, in.sbom, in.scan)
	})
}

func (f *fixture) tryAdd(a actor, in input) error {
	_, err := f.contract.AddRecord(f.opts(a), in.app, in.version, in.scanID, in.artifact, in.sbom, in.scan)
	return err
}

func (f *fixture) recorded(t *testing.T) *fixture {
	t.Helper()
	f.add(f.oracle, defaultRecord())
	return f
}

func (f *fixture) requested(t *testing.T) *types.Receipt {
	t.Helper()
	f.recorded(t)
	return f.send(f.stranger, func(o *bind.TransactOpts) (*types.Transaction, error) {
		return f.contract.RequestVerification(o, big.NewInt(1))
	})
}

func (f *fixture) blockTime(receipt *types.Receipt) uint64 {
	f.t.Helper()
	header, err := f.client.HeaderByHash(context.Background(), receipt.BlockHash)
	if err != nil {
		f.t.Fatal(err)
	}
	return header.Time
}

type dataError interface{ ErrorData() any }

// revert decodes a custom error from a failed call or gas estimation.
func (f *fixture) revert(err error) (string, []any) {
	f.t.Helper()
	if err == nil {
		f.t.Fatal("expected a revert, call succeeded")
	}
	var de dataError
	if !errors.As(err, &de) {
		f.t.Fatalf("error carries no revert data: %v", err)
	}
	text, _ := de.ErrorData().(string)
	data := common.FromHex(text)
	if len(data) < 4 {
		f.t.Fatalf("revert data too short: %q (%v)", text, err)
	}
	for name, e := range f.abi.Errors {
		if bytes.Equal(e.ID[:4], data[:4]) {
			args, uerr := e.Inputs.Unpack(data[4:])
			if uerr != nil {
				f.t.Fatalf("decode %s: %v", name, uerr)
			}
			return name, args
		}
	}
	f.t.Fatalf("unknown custom error selector %x", data[:4])
	return "", nil
}

func sameArg(a, b any) bool {
	if x, ok := a.(*big.Int); ok {
		y, ok := b.(*big.Int)
		return ok && x.Cmp(y) == 0
	}
	return reflect.DeepEqual(a, b)
}

// expectRevert requires a custom error with exactly the given name and arguments.
func (f *fixture) expectRevert(err error, name string, args ...any) {
	f.t.Helper()
	got, gotArgs := f.revert(err)
	if got != name {
		f.t.Fatalf("revert = %s%v, want %s", got, gotArgs, name)
	}
	if len(gotArgs) != len(args) {
		f.t.Fatalf("%s args = %v, want %v", name, gotArgs, args)
	}
	for i := range args {
		if !sameArg(gotArgs[i], args[i]) {
			f.t.Fatalf("%s arg %d = %v, want %v", name, i, gotArgs[i], args[i])
		}
	}
}

func (f *fixture) mustRecord(id int64) registry.AttestationRegistryRecord {
	f.t.Helper()
	r, err := f.contract.GetRecord(nil, big.NewInt(id))
	if err != nil {
		f.t.Fatal(err)
	}
	return r
}

func (f *fixture) count() int64 {
	f.t.Helper()
	n, err := f.contract.RecordCount(nil)
	if err != nil {
		f.t.Fatal(err)
	}
	return n.Int64()
}

func (f *fixture) used(scanID string) bool {
	f.t.Helper()
	u, err := f.contract.IsScanIdUsed(nil, scanID)
	if err != nil {
		f.t.Fatal(err)
	}
	return u
}

func (f *fixture) ids(app, version string, offset, limit *big.Int) []*big.Int {
	f.t.Helper()
	out, err := f.contract.GetRecordIds(nil, app, version, offset, limit)
	if err != nil {
		f.t.Fatal(err)
	}
	return out
}

func (f *fixture) verification(id int64) registry.AttestationRegistryVerificationRequest {
	f.t.Helper()
	v, err := f.contract.GetVerification(nil, big.NewInt(id))
	if err != nil {
		f.t.Fatal(err)
	}
	return v
}

func (f *fixture) isOracle(a common.Address) bool {
	f.t.Helper()
	ok, err := f.contract.IsAuthorizedOracle(nil, a)
	if err != nil {
		f.t.Fatal(err)
	}
	return ok
}

func equalIDs(t *testing.T, got []*big.Int, want ...int64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("ids = %v, want %v", got, want)
	}
	for i := range want {
		if got[i].Int64() != want[i] {
			t.Fatalf("ids = %v, want %v", got, want)
		}
	}
}

func sameRecord(t *testing.T, a, b registry.AttestationRegistryRecord) {
	t.Helper()
	if a.App != b.App || a.Version != b.Version || a.ScanId != b.ScanId ||
		a.ArtifactHash != b.ArtifactHash || a.SbomHash != b.SbomHash || a.ScanHash != b.ScanHash ||
		a.Signer != b.Signer || a.Timestamp.Cmp(b.Timestamp) != 0 {
		t.Fatalf("records differ: %+v vs %+v", a, b)
	}
}

func (f *fixture) authorize(a common.Address) txFunc {
	return func(o *bind.TransactOpts) (*types.Transaction, error) { return f.contract.AuthorizeOracle(o, a) }
}

func (f *fixture) revoke(a common.Address) txFunc {
	return func(o *bind.TransactOpts) (*types.Transaction, error) { return f.contract.RevokeOracle(o, a) }
}

func (f *fixture) request(id *big.Int) txFunc {
	return func(o *bind.TransactOpts) (*types.Transaction, error) { return f.contract.RequestVerification(o, id) }
}

func (f *fixture) fulfill(id *big.Int, status uint8, stale bool) txFunc {
	return func(o *bind.TransactOpts) (*types.Transaction, error) {
		return f.contract.FulfillVerification(o, id, registry.AttestationRegistryVerificationResult{Status: status, Stale: stale})
	}
}

func (f *fixture) try(a actor, fn txFunc) error {
	_, err := fn(f.opts(a))
	return err
}

func appVersionKey(t *testing.T, app, version string) [32]byte {
	t.Helper()
	str, _ := abi.NewType("string", "", nil)
	packed, err := abi.Arguments{{Type: str}, {Type: str}}.Pack(app, version)
	if err != nil {
		t.Fatal(err)
	}
	return crypto.Keccak256Hash(packed)
}

func TestAttestationRegistry(t *testing.T) {
	t.Run("sets the deployer as admin without granting automatic oracle access", func(t *testing.T) {
		f := newFixture(t)
		admin, err := f.contract.Admin(nil)
		if err != nil || admin != f.admin.addr {
			t.Fatalf("admin = %v, %v", admin, err)
		}
		if f.isOracle(f.admin.addr) {
			t.Fatal("admin must not be an oracle by default")
		}
		if !f.isOracle(f.oracle.addr) {
			t.Fatal("oracle should be authorized")
		}
	})

	t.Run("stores exact hashes with the transaction signer and block timestamp", func(t *testing.T) {
		f := newFixture(t)
		receipt := f.add(f.oracle, defaultRecord())
		stored := f.mustRecord(1)
		if stored.App != "demo-lib" || stored.Version != "1.0.0" || stored.ScanId != "scan-1" ||
			stored.ArtifactHash != artifactH || stored.SbomHash != sbomH || stored.ScanHash != scanH {
			t.Fatalf("stored record differs: %+v", stored)
		}
		if stored.Signer != f.oracle.addr {
			t.Fatalf("signer = %v", stored.Signer)
		}
		if stored.Timestamp.Uint64() != f.blockTime(receipt) {
			t.Fatalf("timestamp = %v, want %d", stored.Timestamp, f.blockTime(receipt))
		}
		if f.count() != 1 || !f.used("scan-1") {
			t.Fatal("count or scan ID state wrong")
		}
	})

	t.Run("emits the new record ID and unambiguous app/version key", func(t *testing.T) {
		f := newFixture(t)
		receipt := f.add(f.oracle, defaultRecord())
		if len(receipt.Logs) != 1 {
			t.Fatalf("logs = %d", len(receipt.Logs))
		}
		ev, err := f.contract.ParseRecordAdded(*receipt.Logs[0])
		if err != nil {
			t.Fatal(err)
		}
		if ev.RecordId.Int64() != 1 || ev.AppVersionKey != appVersionKey(t, "demo-lib", "1.0.0") ||
			ev.Signer != f.oracle.addr || ev.App != "demo-lib" || ev.Version != "1.0.0" || ev.ScanId != "scan-1" {
			t.Fatalf("event = %+v", ev)
		}
	})

	t.Run("rejects writes from non-oracles", func(t *testing.T) {
		f := newFixture(t)
		f.expectRevert(f.tryAdd(f.stranger, defaultRecord()), "UnauthorizedOracle", f.stranger.addr)
		if f.count() != 0 || f.used("scan-1") {
			t.Fatal("rejected write changed state")
		}
	})

	for _, operation := range []string{"authorizeOracle", "revokeOracle"} {
		pick := func(f *fixture, a common.Address) txFunc {
			if operation == "authorizeOracle" {
				return f.authorize(a)
			}
			return f.revoke(a)
		}
		t.Run("rejects "+operation+" from non-admins", func(t *testing.T) {
			f := newFixture(t)
			f.expectRevert(f.try(f.stranger, pick(f, f.oracle.addr)), "UnauthorizedAdmin", f.stranger.addr)
		})
		t.Run("rejects the zero address in "+operation, func(t *testing.T) {
			f := newFixture(t)
			f.expectRevert(f.try(f.admin, pick(f, zeroAddress)), "ZeroAddress")
		})
	}

	t.Run("emits authorization and revocation events and rejects redundant changes", func(t *testing.T) {
		f := newFixture(t)
		receipt := f.send(f.admin, f.authorize(f.stranger.addr))
		ev, err := f.contract.ParseOracleAuthorizationChanged(*receipt.Logs[0])
		if err != nil || ev.Oracle != f.stranger.addr || !ev.Authorized {
			t.Fatalf("authorize event = %+v, %v", ev, err)
		}
		f.expectRevert(f.try(f.admin, f.authorize(f.oracle.addr)), "OracleAlreadyAuthorized", f.oracle.addr)
		receipt = f.send(f.admin, f.revoke(f.oracle.addr))
		ev, err = f.contract.ParseOracleAuthorizationChanged(*receipt.Logs[0])
		if err != nil || ev.Oracle != f.oracle.addr || ev.Authorized {
			t.Fatalf("revoke event = %+v, %v", ev, err)
		}
		if f.isOracle(f.oracle.addr) {
			t.Fatal("oracle should be revoked")
		}
		f.expectRevert(f.try(f.admin, f.revoke(f.oracle.addr)), "OracleNotAuthorized", f.oracle.addr)
	})

	fields := []struct {
		name    string
		maximum int
		set     func(*input, string)
	}{
		{"app", maxApp, func(in *input, v string) { in.app = v }},
		{"version", maxVersion, func(in *input, v string) { in.version = v }},
		{"scanId", maxScanID, func(in *input, v string) { in.scanID = v }},
	}
	for _, field := range fields {
		t.Run("rejects an empty "+field.name+" without consuming a record or scan ID", func(t *testing.T) {
			f := newFixture(t)
			in := defaultRecord()
			field.set(&in, "")
			f.expectRevert(f.tryAdd(f.oracle, in), "EmptyField", field.name)
			if f.count() != 0 || f.used("scan-1") {
				t.Fatal("rejected write changed state")
			}
		})
		t.Run("rejects "+field.name+" above its byte limit including UTF-8 input", func(t *testing.T) {
			f := newFixture(t)
			for _, value := range []string{strings.Repeat("a", field.maximum+1), strings.Repeat("é", field.maximum/2+1)} {
				in := defaultRecord()
				field.set(&in, value)
				f.expectRevert(f.tryAdd(f.oracle, in), "StringTooLong",
					field.name, big.NewInt(int64(len([]byte(value)))), big.NewInt(int64(field.maximum)))
			}
		})
	}

	t.Run("accepts strings exactly at the byte limits", func(t *testing.T) {
		f := newFixture(t)
		in := input{
			app: strings.Repeat("a", maxApp), version: strings.Repeat("v", maxVersion), scanID: strings.Repeat("s", maxScanID),
			artifact: artifactH, sbom: sbomH, scan: scanH,
		}
		f.add(f.oracle, in)
		stored := f.mustRecord(1)
		if stored.App != in.app || stored.Version != in.version || stored.ScanId != in.scanID {
			t.Fatal("stored strings differ")
		}
	})

	hashFields := []struct {
		name string
		set  func(*input)
	}{
		{"artifactHash", func(in *input) { in.artifact = zeroHash }},
		{"sbomHash", func(in *input) { in.sbom = zeroHash }},
		{"scanHash", func(in *input) { in.scan = zeroHash }},
	}
	for _, field := range hashFields {
		t.Run("rejects a zero "+field.name, func(t *testing.T) {
			f := newFixture(t)
			in := defaultRecord()
			field.set(&in)
			f.expectRevert(f.tryAdd(f.oracle, in), "EmptyHash", field.name)
		})
	}

	t.Run("rejects scan ID replay across applications, versions, and oracles", func(t *testing.T) {
		f := newFixture(t)
		f.recorded(t)
		f.send(f.admin, f.authorize(f.stranger.addr))
		f.expectRevert(f.tryAdd(f.stranger, record("scan-1", "other-app", "2.0.0")), "ScanIdAlreadyUsed", "scan-1")
		if f.count() != 1 {
			t.Fatal("replay was recorded")
		}
	})

	t.Run("preserves records and used scan IDs after revocation and reauthorization", func(t *testing.T) {
		f := newFixture(t)
		f.recorded(t)
		original := f.mustRecord(1)
		f.send(f.admin, f.revoke(f.oracle.addr))
		f.expectRevert(f.tryAdd(f.oracle, record("scan-2", "demo-lib", "1.0.0")), "UnauthorizedOracle", f.oracle.addr)
		sameRecord(t, f.mustRecord(1), original)
		if !f.used("scan-1") {
			t.Fatal("scan ID was released by revocation")
		}
		f.send(f.admin, f.authorize(f.oracle.addr))
		f.expectRevert(f.tryAdd(f.oracle, defaultRecord()), "ScanIdAlreadyUsed", "scan-1")
	})

	t.Run("appends rescans and changes latest without modifying older records", func(t *testing.T) {
		f := newFixture(t)
		f.recorded(t)
		original := f.mustRecord(1)
		f.add(f.oracle, record("scan-2", "demo-lib", "1.0.0"))
		latest, err := f.contract.GetLatestRecord(nil, "demo-lib", "1.0.0")
		if err != nil || latest.ScanId != "scan-2" {
			t.Fatalf("latest = %+v, %v", latest, err)
		}
		sameRecord(t, f.mustRecord(1), original)
		equalIDs(t, f.ids("demo-lib", "1.0.0", big.NewInt(0), big.NewInt(100)), 1, 2)
	})

	t.Run("keeps ambiguous concatenations and different versions in separate histories", func(t *testing.T) {
		f := newFixture(t)
		f.recorded(t)
		f.add(f.oracle, record("scan-2", "ab", "c"))
		f.add(f.oracle, record("scan-3", "a", "bc"))
		f.add(f.oracle, record("scan-4", "demo-lib", "2.0.0"))
		for _, c := range []struct{ app, version, scan string }{
			{"ab", "c", "scan-2"}, {"a", "bc", "scan-3"}, {"demo-lib", "1.0.0", "scan-1"},
		} {
			latest, err := f.contract.GetLatestRecord(nil, c.app, c.version)
			if err != nil || latest.ScanId != c.scan {
				t.Fatalf("latest(%s,%s) = %+v, %v", c.app, c.version, latest, err)
			}
		}
		equalIDs(t, f.ids("demo-lib", "2.0.0", big.NewInt(0), big.NewInt(100)), 4)
	})

	t.Run("paginates oldest first and handles empty history and extreme offsets", func(t *testing.T) {
		f := newFixture(t)
		f.recorded(t)
		f.add(f.oracle, record("scan-2", "demo-lib", "1.0.0"))
		f.add(f.oracle, record("scan-3", "demo-lib", "1.0.0"))
		equalIDs(t, f.ids("demo-lib", "1.0.0", big.NewInt(0), big.NewInt(2)), 1, 2)
		equalIDs(t, f.ids("demo-lib", "1.0.0", big.NewInt(2), big.NewInt(2)), 3)
		for _, offset := range []*big.Int{big.NewInt(3), big.NewInt(4), maxUint256} {
			equalIDs(t, f.ids("demo-lib", "1.0.0", offset, big.NewInt(100)))
		}
		equalIDs(t, f.ids("missing", "1.0.0", big.NewInt(0), big.NewInt(100)))
	})

	t.Run("rejects zero and oversized page limits", func(t *testing.T) {
		f := newFixture(t)
		f.recorded(t)
		for _, limit := range []*big.Int{big.NewInt(0), big.NewInt(101), maxUint256} {
			_, err := f.contract.GetRecordIds(nil, "demo-lib", "1.0.0", big.NewInt(0), limit)
			f.expectRevert(err, "InvalidPageLimit", limit)
		}
	})

	t.Run("rejects missing records and missing latest history", func(t *testing.T) {
		f := newFixture(t)
		f.recorded(t)
		for _, id := range []*big.Int{big.NewInt(0), big.NewInt(2), maxUint256} {
			_, err := f.contract.GetRecord(nil, id)
			f.expectRevert(err, "RecordNotFound", id)
			f.expectRevert(f.try(f.stranger, f.request(id)), "RecordNotFound", id)
		}
		_, err := f.contract.GetLatestRecord(nil, "missing", "1.0.0")
		f.expectRevert(err, "NoRecordsForAppVersion", "missing", "1.0.0")
		n, err := f.contract.VerificationRequestCount(nil)
		if err != nil || n.Sign() != 0 {
			t.Fatalf("verification count = %v, %v", n, err)
		}
	})

	t.Run("creates a pending request with the requester and block time and emits an event", func(t *testing.T) {
		f := newFixture(t)
		receipt := f.requested(t)
		request := f.verification(1)
		if request.RecordId.Int64() != 1 || request.Requester != f.stranger.addr {
			t.Fatalf("request = %+v", request)
		}
		if request.RequestedAt.Uint64() != f.blockTime(receipt) {
			t.Fatalf("requestedAt = %v", request.RequestedAt)
		}
		if request.Fulfilled || request.Result.Status != 0 || request.Result.Stale ||
			request.FulfilledBy != zeroAddress || request.FulfilledAt.Sign() != 0 {
			t.Fatalf("request not pending: %+v", request)
		}
		ev, err := f.contract.ParseVerificationRequested(*receipt.Logs[0])
		if err != nil || ev.RequestId.Int64() != 1 || ev.RecordId.Int64() != 1 || ev.Requester != f.stranger.addr {
			t.Fatalf("event = %+v, %v", ev, err)
		}
	})

	t.Run("fulfills once with oracle identity and block time without changing the record", func(t *testing.T) {
		f := newFixture(t)
		f.requested(t)
		original := f.mustRecord(1)
		receipt := f.send(f.oracle, f.fulfill(big.NewInt(1), 1, true))
		response := f.verification(1)
		if !response.Fulfilled || response.FulfilledBy != f.oracle.addr ||
			response.FulfilledAt.Uint64() != f.blockTime(receipt) ||
			response.Result.Status != 1 || !response.Result.Stale {
			t.Fatalf("response = %+v", response)
		}
		sameRecord(t, f.mustRecord(1), original)
		ev, err := f.contract.ParseVerificationFulfilled(*receipt.Logs[0])
		if err != nil || ev.RequestId.Int64() != 1 || ev.RecordId.Int64() != 1 || ev.Oracle != f.oracle.addr ||
			ev.Status != 1 || !ev.Stale {
			t.Fatalf("event = %+v, %v", ev, err)
		}
		f.expectRevert(f.try(f.oracle, f.fulfill(big.NewInt(1), 1, true)), "VerificationAlreadyFulfilled", big.NewInt(1))
	})

	t.Run("rejects fulfillment by outsiders and revoked oracles", func(t *testing.T) {
		f := newFixture(t)
		f.requested(t)
		f.expectRevert(f.try(f.stranger, f.fulfill(big.NewInt(1), 1, false)), "UnauthorizedOracle", f.stranger.addr)
		f.send(f.admin, f.revoke(f.oracle.addr))
		f.expectRevert(f.try(f.oracle, f.fulfill(big.NewInt(1), 1, false)), "UnauthorizedOracle", f.oracle.addr)
		if f.verification(1).Fulfilled {
			t.Fatal("request was fulfilled")
		}
	})

	t.Run("rejects missing verification requests", func(t *testing.T) {
		f := newFixture(t)
		f.requested(t)
		for _, id := range []*big.Int{big.NewInt(0), big.NewInt(2), maxUint256} {
			_, err := f.contract.GetVerification(nil, id)
			f.expectRevert(err, "VerificationRequestNotFound", id)
			f.expectRevert(f.try(f.oracle, f.fulfill(id, 1, false)), "VerificationRequestNotFound", id)
		}
	})

	t.Run("rejects Pending results and stale flags on non-valid results", func(t *testing.T) {
		f := newFixture(t)
		f.requested(t)
		type result struct {
			status uint8
			stale  bool
		}
		results := []result{{0, false}}
		for status := uint8(2); status <= 7; status++ {
			results = append(results, result{status, true})
		}
		for _, r := range results {
			f.expectRevert(f.try(f.oracle, f.fulfill(big.NewInt(1), r.status, r.stale)), "InvalidVerificationResult")
		}
		if f.verification(1).Fulfilled {
			t.Fatal("request was fulfilled")
		}
	})

	t.Run("supports all final integrity results on independent requests", func(t *testing.T) {
		f := newFixture(t)
		f.recorded(t)
		for status := uint8(1); status <= 7; status++ {
			f.send(f.stranger, f.request(big.NewInt(1)))
			n, err := f.contract.VerificationRequestCount(nil)
			if err != nil {
				t.Fatal(err)
			}
			f.send(f.oracle, f.fulfill(n, status, false))
			got := f.verification(n.Int64()).Result
			if got.Status != status || got.Stale {
				t.Fatalf("result = %+v, want status %d", got, status)
			}
		}
	})
}
