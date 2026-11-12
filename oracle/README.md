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
