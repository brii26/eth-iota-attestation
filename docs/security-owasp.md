# Attestation registry security review

This review covers `smart-contract/contracts/AttestationRegistry.sol` and its
test suite in `smart-contract/registry/registry_test.go`. It assesses the
current registry behavior, including its trust assumptions and remaining risks.
It is an implementation review, without a claim of independent audit.

The reference is the [OWASP Smart Contract Top 10, 2026 edition](https://scs.owasp.org/sctop10/).
The 2026 list places business logic at SC02 and input validation at SC05.
Denial of service and insecure randomness are considered separately below.

## Mitigations and test evidence

The names in the last column are the exact strings passed to `t.Run` in
`TestAttestationRegistry`. Run them with `cd smart-contract && go test ./registry`.

| Risk | Implementation and remaining limits | Test evidence |
| --- | --- | --- |
| Access control | `onlyAdmin` gates permission changes. `onlyOracle` checks current authorization for recording and fulfillment. Deployment does not authorize the administrator automatically. | `sets the deployer as admin without granting automatic oracle access`, `rejects writes from non-oracles`, `rejects authorizeOracle from non-admins`, `rejects revokeOracle from non-admins` |
| Permission changes | Permission changes emit `OracleAuthorizationChanged`. Redundant changes and zero addresses revert. Revocation blocks future writes, while existing records remain available. | `emits authorization and revocation events and rejects redundant changes`, `rejects the zero address in authorizeOracle`, `rejects the zero address in revokeOracle`, `preserves records and used scan IDs after revocation and reauthorization`, `rejects fulfillment by outsiders and revoked oracles` |
| Empty or oversized input | Required strings and hashes must be nonempty. Limits apply to UTF-8 bytes, including multibyte input. Semantic package/version validation remains the oracle's responsibility. | `rejects an empty app without consuming a record or scan ID`, `rejects an empty version without consuming a record or scan ID`, `rejects an empty scanId without consuming a record or scan ID`, `rejects app above its byte limit including UTF-8 input`, `rejects version above its byte limit including UTF-8 input`, `rejects scanId above its byte limit including UTF-8 input`, `accepts strings exactly at the byte limits`, `rejects a zero artifactHash`, `rejects a zero sbomHash`, `rejects a zero scanHash` |
| Scan replay | Scan IDs are unique across the deployment and remain consumed after oracle revocation. A new scan ID for old content is not detected by this rule. | `rejects scan ID replay across applications, versions, and oracles`, `preserves records and used scan IDs after revocation and reauthorization` |
| Record integrity and indexing | Records have no update or delete method. Signer and timestamp come from the transaction. App/version keys use `abi.encode` to avoid ambiguous concatenation. Latest follows recording order, not claimed scan time. | `stores exact hashes with the transaction signer and block timestamp`, `emits the new record ID and unambiguous app/version key`, `appends rescans and changes latest without modifying older records`, `keeps ambiguous concatenations and different versions in separate histories` |
| Missing IDs | Read and verification methods reject zero and unknown IDs instead of returning a default struct. Missing history returns an empty page. | `rejects missing records and missing latest history`, `rejects missing verification requests`, `paginates oldest first and handles empty history and extreme offsets` |
| Verification lifecycle | Anyone may request verification of an existing record. Only a current oracle may fulfill once. Pending cannot be a final result and stale is allowed only with Valid. | `creates a pending request with the requester and block time and emits an event`, `fulfills once with oracle identity and block time without changing the record`, `rejects fulfillment by outsiders and revoked oracles`, `rejects Pending results and stale flags on non-valid results`, `supports all final integrity results on independent requests` |
| Bounded work and pagination | Write functions contain no source-level loops. Read loops return at most 100 IDs. Offset handling avoids adding an attacker-chosen limit to an extreme offset. Public verification requests can still grow storage. | `paginates oldest first and handles empty history and extreme offsets`, `rejects zero and oversized page limits` |
| Oracle manipulation | The registry records the supplied hashes and conclusions. Authorization cannot establish that files, scans, or outcomes are truthful. Revocation stops future actions without rewriting old records. | `stores exact hashes with the transaction signer and block timestamp`, `preserves records and used scan IDs after revocation and reauthorization`, `rejects fulfillment by outsiders and revoked oracles` |

## Coverage of the 2026 categories

| OWASP category | Assessment |
| --- | --- |
| SC01 Access control | Covered by administrator and oracle checks, with explicit permission-change events and negative tests above. A stolen administrator key can authorize a malicious oracle. |
| SC02 Business logic | Append-only history, global scan IDs, app/version separation, and one-time fulfillment are covered by the invariant tests above. |
| SC03 Price oracle manipulation | No price feed or asset pricing exists. Attestation-oracle manipulation remains a separate trust risk described below. |
| SC04 Flash loan facilitated attacks | No borrowing, balances, collateral, rewards, or externally priced operation exists in the registry. |
| SC05 Input validation | Bounds, nonzero hashes, identifiers, missing records, and result combinations are checked and tested above. |
| SC06 Unchecked external calls | Contract source contains no external calls. No external-call return values need handling. |
| SC07 Arithmetic errors | There are no rates, ratios, scaling calculations, or value conversions. IDs and pagination use checked integer operations. Extreme pagination cases are tested above. |
| SC08 Reentrancy | No external call or value transfer provides a callback path. This assessment follows code inspection, rather than an exploit test. |
| SC09 Integer overflow and underflow | Solidity 0.8 checked arithmetic is enabled. Pagination checks the offset before subtracting it from history length. Extreme offsets and limits are tested above. |
| SC10 Proxy and upgradeability | Deployment is direct, without a proxy, initializer, or upgrade method. Registry replacement requires a new deployment and explicit configuration changes. |

Insecure randomness is not used. The block timestamp records when a transaction
was included, without serving as randomness, a precise scan time, or proof of
freshness. Denial of service remains relevant even though it is outside the
numbered 2026 list. Bounded functions do not prevent an authorized oracle or
public requester from creating many transactions and growing storage.

## Oracle trust and operational risks

The oracle is a trusted component. A compromised oracle can record arbitrary
nonzero hashes, fabricate results, or race another oracle to consume a scan ID.
The existing tests show that the registry preserves supplied bytes and enforces
authorization. They do not prove that the external data is true.

The backend design mitigates some risks by authenticating pipelines, downloading
artifacts independently, hashing original file bytes, checking claimed hashes,
and validating Grype database freshness. These are backend requirements that
are not implemented by this contract and are not proven by its unit tests.

Current oracle authorization must be checked when verifying historical records
and responses. Revocation cannot undo an action already included before the
revocation transaction. Submitted hashes do not prove the identity of a scanner
or the safety of an artifact. Unavailable files also prevent independent
verification even when the on-chain record remains intact.

The administrator is immutable. Losing that key prevents future permission
changes, and compromise requires a replacement registry and a migration plan.
Private keys must be kept out of the repository. Public verification requests
need gas on a normal chain, but there is no contract-level request rate limit
or response deadline. A failed or revoked oracle can leave requests pending.

## Verification result

All 32 contract tests pass with Solidity 0.8.28 and the Cancun EVM target.
The test evidence covers the behaviors
listed above. It does not establish the absence of every possible vulnerability
or validate the future backend, wallet UI, Geth deployment, or IOTA adapter.
