# Attestation registry

`AttestationRegistry` stores software release attestations on Ethereum and
coordinates verification requests. Deploying it makes the deployer the immutable
administrator. The administrator must explicitly authorize oracle addresses
before they can record or fulfill anything.

## Records

`addRecord(app, version, scanId, artifactHash, sbomHash, scanHash)` returns a new
record ID. IDs start at one and increase in recording order. The contract stores
the supplied fields together with `msg.sender` as `signer` and `block.timestamp`
as `timestamp`. The three hashes are nonzero `bytes32` SHA-256 fingerprints of
the original files. The contract cannot check whether the supplied hashes match
the files or whether the scan is correct.

| Field | Maximum UTF-8 bytes |
| --- | --- |
| `app` | 214 |
| `version` | 128 |
| `scanId` | 128 |

Strings must be nonempty. Identity comparisons use their exact bytes, without
normalization. Validation of package naming and version formats belongs to the
oracle. A scan ID is unique across all applications, versions, and oracle
addresses in this registry deployment. Revoking an oracle does not release its
scan IDs or change its historical records.

`getRecord(id)` and `getLatestRecord(app, version)` return a record or revert
with a specific not-found error. `getRecordIds(app, version, offset, limit)`
returns IDs oldest first, with a limit from 1 to `MAX_PAGE_SIZE` (100). Missing
history and offsets at or beyond the end return an empty array. Applications
and versions are indexed with `keccak256(abi.encode(app, version))`.

There is no update or delete operation. `recordCount` reports the number of
stored records, and `isScanIdUsed(scanId)` checks replay protection.

## Permissions

Only `admin` can call `authorizeOracle(address)` and `revokeOracle(address)`.
Zero addresses and redundant permission changes revert.
`isAuthorizedOracle(address)` reports the current permission. Revoked oracles
cannot write new records or fulfill requests. A verifier must check current
authorization when deciding whether to trust an existing record or response.

## Verification requests

Anyone can call `requestVerification(recordId)` for an existing record. It
returns a request ID, starting at one. `getVerification(requestId)` returns the
record ID, requester, request time, fulfillment flag, result, responding oracle,
and response time. Missing requests revert. A pending request has
`fulfilled = false` and a `Pending` result.

An authorized oracle calls `fulfillVerification(requestId, result)` once with
the tuple `{ status, stale }`. The result status is an enum:

| Value | Status |
| --- | --- |
| 0 | `Pending` |
| 1 | `Valid` |
| 2 | `ArtifactMismatch` |
| 3 | `SbomMismatch` |
| 4 | `ScanMismatch` |
| 5 | `UntrustedSigner` |
| 6 | `RecordNotFound` |
| 7 | `FileUnavailable` |

`Pending` cannot be submitted as a final result. `stale` may be true only with
`Valid`, keeping freshness separate from integrity. The responding oracle and
timestamp come from the transaction. Fulfillment does not change the original
attestation. The contract records the oracle's conclusion and does not fetch or
verify off-chain data itself.

## Events

| Event | Purpose |
| --- | --- |
| `OracleAuthorizationChanged` | Announce an oracle permission change |
| `RecordAdded` | Identify a new record, application/version key, signer, and scan |
| `VerificationRequested` | Identify a request, its record, and requester |
| `VerificationFulfilled` | Identify the responding oracle and result |

Read the full record or request through its getter using the event ID.
Events index the fields needed for record lookup and request processing.
Write operations have no external calls or loops. History reads are bounded
by the page-size limit.
