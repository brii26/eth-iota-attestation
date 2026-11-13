# IOTA notarization spike

This spike investigates locked notarization for artifact attestations. It stays
in TypeScript because the notarization library exists for Rust and WebAssembly
with TypeScript bindings and has no Go version. It creates disposable test records, checks their lifecycle,
and exports evidence without storing signing keys. It is restricted to testnet or localnet.

## Investigated Findings & Evidence

The spike script was verified against the pinned SDK (`@iota/notarization@0.1.15`, `@iota/iota-sdk@1.17.0`, and `@iota/iota-interaction-ts@0.14.0`) and the official IOTA testnet deployment.

| Question | Documented behavior | Verified Behavior / Evidence |
| --- | --- | --- |
| **Create and read** | A locked object holds immutable state with owner and creation metadata | Confirmed. `NotarizationClient.createLocked()` writes immutable state and `NotarizationClientReadOnly.getNotarizationById()` retrieves `method: "Locked"`, state payload, owner address, and metadata. Package ID auto-resolves to `0x00412bd469b7f980227c6c574090348239852e43aa07818b315854fdd8a2d25f` on testnet. |
| **Update** | Locked state and metadata cannot be changed | Confirmed rejected. `updateState` on a locked notarization aborts with `MoveAbort` (`UntilDestroyed` update lock prevents state mutation). |
| **Destroy before expiry** | A future delete lock prevents destruction | Confirmed rejected. Calling `destroy()` before the delete-lock timestamp aborts on-chain. `readOnly.isDestroyAllowed(id)` evaluates to `false`. |
| **Destroy after expiry** | The owner can destroy the object after expiry | Confirmed permitted. After ledger timestamp exceeds `unlockAt`, `readOnly.isDestroyAllowed(id)` evaluates to `true` and `destroy()` destroys the object, releasing the storage rebate. |
| **No delete lock** | Destruction is immediately allowed | Confirmed permitted. A locked notarization without a delete lock allows immediate owner destruction. |
| **`UntilDestroyed` delete lock** | This configuration is invalid | Confirmed rejected. Attempting to assign `UntilDestroyed` to `withDeleteLock` fails client-side / Move validation. |
| **Sender and timestamp** | Transaction sender and notarization creation metadata are available | Confirmed. Transaction sender is extracted from `response.transaction.data.sender`, and recording timestamp is verified from `readOnly.createdAtTs(id)` matching ledger checkpoint time. |
| **Transaction status and cost** | Effects provide status and computation, storage, and rebate fields | Confirmed. `effects.gasUsed` breaks down `computationCost`, `storageCost`, and `storageRebate`. Net gas equals computation + storage - rebate. |
| **Owned objects** | The SDK supports paginated owner queries | Confirmed. `rpc.getOwnedObjects({ owner, cursor, limit })` returns paginated object IDs. |
| **Parallel account writes** | Transactions sharing an owned gas object may conflict | Confirmed. Concurrent transactions from the same account contend for the same unspent gas coin, resulting in `object version unavailable` or `gas object locked` unless distinct gas coins or serial queues are used. |

> [!IMPORTANT]
> **Locked does not mean permanently available.** The state remains immutable while
> the object exists, but permitted destruction can remove it. The oracle application
> must record the delete-lock horizon and treat object disappearance as a
> verification availability failure (`FILE_UNAVAILABLE`).

Sources:
- [Create a locked notarization](https://docs.iota.org/developer/iota-notarization/single-notarization/how-tos/locked-notarizations/create)
- [Destroy a locked notarization](https://docs.iota.org/developer/iota-notarization/single-notarization/how-tos/locked-notarizations/destroy)
- [Notarization client API](https://docs.iota.org/developer/iota-notarization/single-notarization/references/wasm/notarization_wasm/classes/NotarizationClient)
- [Owned-object and gas contention](https://docs.iota.org/developer/iota-101/transactions/sponsored-transactions/about-sponsored-transactions)

## Install and configure

From `smart-contract/iota/`:

```sh
cd smart-contract/iota
npm install
cp .env.example .env
chmod 600 .env
```

The manifest pins `@iota/notarization` to `0.1.15`, `@iota/iota-sdk` to `1.17.0`,
and `@iota/iota-interaction-ts` to `0.14.0`. The `package-lock.json` lockfile is committed
to ensure reproducible dependency resolution.

### Configuration (`.env`)

- `IOTA_NETWORK`: `testnet` or `localnet` (default: `testnet`).
- `IOTA_RPC_URL`: RPC endpoint (default: `https://api.testnet.iota.cafe`).
- `IOTA_NOTARIZATION_PKG_ID`: Optional. If omitted, the client auto-resolves the default deployed package on testnet (`0x00412bd469b7f980227c6c574090348239852e43aa07818b315854fdd8a2d25f`).
- `IOTA_PRIVATE_KEY`: Optional 32-byte Ed25519 hex key or encoded private key. If left empty, an ephemeral keypair is generated.
- `IOTA_REQUEST_FAUCET`: Set to `1` when running against a localnet node with programmatic faucet support. (For testnet, request tokens via web faucet at `https://faucet.testnet.iota.cafe`).
- `IOTA_LOCK_SECONDS`: Delete-lock duration in seconds (default: `120`).
- `IOTA_WAIT_FOR_UNLOCK`: Set to `1` to wait for lock expiration and verify destroy-after-unlock.
- `IOTA_TEST_PARALLEL`: Set to `1` to probe gas coin contention with parallel writes.

## Run

From `smart-contract/iota/`:

1. **Typecheck TypeScript types**:
   ```sh
   npm run typecheck
   ```

2. **Offline payload validation test**:
   ```sh
   npm test
   ```

3. **Execute on-chain spike**:
   ```sh
   npm run spike
   ```

## Evidence and failure handling

Each run exports results to `smart-contract/iota/results/<run-id>.json` (git-ignored).
It logs:
- Object IDs of created notarizations.
- Confirmed transaction digests, effects, and execution statuses.
- Gas usage components (computation, storage, storage rebate, and net cost).
- MoveAbort and validation error captures for rejected operations.
- Owned object list verification.

## Record Format Decision

For the Oracle IOTA Adapter (Issue #11):

1. **State Payload**:
   Store one UTF-8 JSON value per notarization object formatted as:
   ```json
   {
     "schemaVersion": 1,
     "app": "demo-lib",
     "version": "1.4.2",
     "scanId": "0f8c2d3e-6b1a-4c55-9e1f-2a7b8c9d0e11",
     "artifactHash": "0x466357abeed49b9316d407eda70b16fb7185f55011932bc0b2c50bf058091a14",
     "sbomHash": "0xf7a1c21f3860a14cc966fbf31c574a4b1fa58510341e015790901e46e33db11f",
     "scanHash": "0xba7ff3dabf002a9827bf0a6b1e352b8ffd259fb4437e4bb614cf83bed060f596"
   }
   ```
2. **Metadata & Signer**:
   - The signer address is derived directly from the transaction sender (`effects.transaction.data.sender`).
   - The creation timestamp is derived from ledger metadata (`readOnly.createdAtTs(objectId)`).
3. **Off-chain Indexing (SQLite)**:
   - Since IOTA locked notarizations are unindexed independent objects, the Oracle SQLite index must maintain:
     - `(network, scan_id)` uniqueness.
     - `(network, app, version)` pointer to latest `objectId` and `txDigest`.
   - The Oracle periodically reconciles indexed object IDs with on-chain object existence (`rpc.getObject`) to detect if an object has been destroyed or pruned.
4. **Gas Coin Concurrency Control**:
   - Because IOTA transactions lock owned gas coins during execution, the Oracle adapter must serialize transactions from a single signer or maintain a pool of separate gas coin objects to avoid parallel contention errors.
