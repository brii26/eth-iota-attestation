# Oracle and backend

Reserved for the Fastify and TypeScript API, oracle, and chain adapters.
The backend and oracle share this directory. A separate backend service is
not required for the proof of concept.

The oracle will validate submissions, fetch artifacts from Verdaccio, hash the
original bytes, store SBOMs and scans in RustFS, and maintain a rebuildable
SQLite index. Chain adapters will support Ethereum and IOTA, with network
selection per request. The verification API will retrieve files and compare
their hashes against the selected on-chain record.

See the [project overview](../README.md) for the system architecture.
