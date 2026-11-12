# IOTA notarization spike

This spike investigates locked notarization for artifact attestations. It stays
in TypeScript because the notarization library exists for Rust and WebAssembly
with TypeScript bindings and has no Go version. It creates disposable test records, checks their lifecycle,
and exports evidence without storing signing keys. It is restricted to testnet or localnet.

## Current evidence

The script has not been executed against an IOTA network. Dependency download
and network access are required before SDK typechecking and on-chain execution.
No object IDs, transaction digests, measured costs, or concurrency outcomes are
claimed in this document. The following conclusions come from official
documentation and must be checked against the pinned SDK and deployed package.

| Question | Documented behavior | Runtime evidence |
| --- | --- | --- |
| Create and read | A locked object holds immutable state with owner and creation metadata | Pending |
| Update | Locked state and metadata cannot be changed | Pending rejection and unchanged-state check |
| Destroy before expiry | A future delete lock prevents destruction | Pending rejection |
| Destroy after expiry | The owner can destroy the object after expiry | Pending transaction |
| No delete lock | Destruction is immediately allowed | Pending transaction |
| `UntilDestroyed` delete lock | This configuration is invalid | Pending SDK or Move rejection |
| Sender and timestamp | Transaction sender and notarization creation metadata are available | Pending comparison with ledger clock |
| Transaction status and cost | Effects provide status and computation, storage, and rebate fields | Pending transaction effects |
| Owned objects | The SDK supports paginated owner queries | Pending object listing |
| Parallel account writes | Transactions sharing an owned gas object may conflict | Pending two-write probe |

Locked does not mean permanently available. The state remains immutable while
the object exists, but permitted destruction can remove it. The application
must record the delete-lock horizon and treat object disappearance as a
verification availability failure.

Sources:

- [Create a locked notarization](https://docs.iota.org/developer/iota-notarization/single-notarization/how-tos/locked-notarizations/create)
- [Destroy a locked notarization](https://docs.iota.org/developer/iota-notarization/single-notarization/how-tos/locked-notarizations/destroy)
- [Notarization client API](https://docs.iota.org/developer/iota-notarization/single-notarization/references/wasm/notarization_wasm/classes/NotarizationClient)
- [Owned-object and gas contention](https://docs.iota.org/developer/iota-101/transactions/sponsored-transactions/about-sponsored-transactions)

## Install and configure

From `smart-contract/`, first install the parent tooling with `npm ci`. Then
install the isolated spike dependencies:

```sh
cd iota
npm install
cp .env.example .env
chmod 600 .env
```

The manifest pins `@iota/notarization` to 0.1.15, `@iota/iota-sdk` to 1.17.0,
and its signer helper `@iota/iota-interaction-ts` to 0.14.0. A dependency lockfile
has not been generated yet because package installation could not complete.
Resolve and commit that lockfile after a successful install before considering
the spike reproducible. See [the official Wasm setup](https://docs.iota.org/developer/iota-notarization/single-notarization/getting-started/wasm).

Set `IOTA_NOTARIZATION_PKG_ID` to a 32-byte package ID deployed on the selected
network. Use the [network setup guide](https://docs.iota.org/developer/iota-notarization/single-notarization/getting-started/local-network-setup)
to deploy a matching package on localnet when a testnet package is unavailable.
Do not use a package ID from a different network.

The default configuration requests faucet funds for a fresh ephemeral account.
To use a funded account instead, set `IOTA_PRIVATE_KEY` to its Ed25519 raw hex
or encoded secret key and set `IOTA_REQUEST_FAUCET=0`. The key stays in the
environment. A fresh account isolates gas-object contention from other work.

## Run

From `smart-contract/iota/`:

```sh
npm run typecheck
npm run spike
```

The default delete lock lasts 120 seconds. The script waits for the ledger
clock to pass expiry, destroys the temporary object, then probes two concurrent
writes from the same account. These writes can contend for a gas coin. Set
`IOTA_TEST_PARALLEL=0` to skip that probe or `IOTA_WAIT_FOR_UNLOCK=0` to skip the
wait. A skipped scenario leaves the report incomplete.
Incomplete runs exit with code 2 and execution failures exit with code 1.

For an SDK-free payload check from `smart-contract/`:

```sh
node --import tsx iota/notarization-spike.ts --check
```

This check does not contact the network or validate the SDK API. It cannot
satisfy the on-chain spike acceptance criteria.

## Evidence and failure handling

Each run saves `iota/results/<run-id>.json`, an ignored generated artifact.
It includes object IDs, confirmed transaction digests, effects, sender,
timestamps, gas components, lock observations, owned objects, and parallel
attempt outcomes when those operations execute. Errors and partial results
are preserved if a run fails. Copy confirmed findings into this document
without copying signing keys.

A network error, missing gas, or invalid signature is not accepted as proof of
a lock rejection. Rejection evidence records whether it came from a Move abort
or client-side delete-lock validation. Concurrent failures are recorded as
observations rather than assumed to be gas conflicts. Inspect each cause
before drawing a conclusion. If an ephemeral run stops early, the remaining
objects may be left behind because its signing key is not saved.

## Record format decision

Store one UTF-8 JSON value per scan with fields in this order:
`schemaVersion`, `app`, `version`, `scanId`, `artifactHash`, `sbomHash`, `scanHash`.
Use schema version 1 and 0x-prefixed lowercase SHA-256 hex strings for hashes.
The original artifact, SBOM, and scan stay off-chain. The JSON contains no
user-supplied signer or recording timestamp.

Read signer from the confirmed transaction sender and time from the ledger
creation metadata. The script checks the latter against the checkpoint clock
before treating it as milliseconds. Keep object ID, creation digest, package
ID, and delete-lock expiry in the off-chain index. These mappings remain
provisional until the network run succeeds.

The oracle and index must enforce scan ID uniqueness and select the latest
record for an app/version. Locked notarization alone does not enforce either
rule across separate objects. A successful two-write probe also does not
establish that concurrent writes from one account are safe at every load.
