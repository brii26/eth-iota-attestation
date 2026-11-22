# Oracle and backend

Reserved for the Go API, oracle, and chain adapters. The backend and oracle
share this directory and the Go module at the repository root. A separate
backend service is not required for the proof of concept.

The API uses the standard library `net/http` router. Planned packages:

| Path | Responsibility |
| --- | --- |
| `cmd/oracle/` | Service entry point |
| `internal/api/` | Handlers, middleware, and error responses |
| `internal/chain/` | `writeRecord` and `readRecord` interface and the Ethereum adapter |
| `internal/store/` | SQLite index and RustFS access |
| `internal/verify/` | Hashing and verification statuses |

The oracle will validate submissions, fetch artifacts from Verdaccio, hash the
original bytes, store SBOMs and scans in RustFS, and maintain a rebuildable
SQLite index. Chain adapters will support Ethereum and IOTA, with network
selection per request. The Ethereum adapter uses the generated bindings in
`smart-contract/registry`. The IOTA adapter is planned as a small TypeScript
service because the notarization library has no Go version. The verification API
will retrieve files and compare their hashes against the selected on-chain record.

See the [project overview](../README.md) for the system architecture.

## GitHub Actions OIDC

Submissions require a GitHub Actions ID token. Set `OIDC_AUDIENCE` to an
application-specific audience and set `OIDC_APP_ALLOWLIST` to a JSON map
from package names to their permitted repository and workflow ref:

```json
{
  "@attestation/demo": {
    "repository": "brii26/eth-iota-attestation",
    "workflowRef": "brii26/eth-iota-attestation/.github/workflows/attest.yml@refs/heads/main"
  }
}
```

Configure repository variables `OIDC_AUDIENCE`, `ORACLE_URL`,
`VERDACCIO_URL`, and optionally `ATTESTATION_NETWORK`. Configure secrets
`TS_OAUTH_CLIENT_ID`, `TS_OAUTH_SECRET`, and `VERDACCIO_TOKEN`. The registry
and oracle URLs must be reachable from the runner through the tailnet. Expose
only their application ports to `tag:ci` in the tailnet policy. The oracle
host needs outbound HTTPS to GitHub's OIDC issuer for signing keys.

The workflow packs the example package with a unique prerelease version,
publishes to Verdaccio, creates CycloneDX and Grype reports, and submits the
original report bytes using the base64 JSON fields. The existing embedded
JSON fields remain supported. Use only one representation per request.
