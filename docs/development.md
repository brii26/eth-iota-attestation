# Development setup

## Repository layout

| Path | Responsibility |
| --- | --- |
| `frontend/` | React dashboard and MetaMask integration |
| `smart-contract/` | Solidity contract, Hardhat 3 compiler, Go bindings, tests, and private chain commands |
| `oracle/` | Go backend, oracle, chain adapters, and SQLite index |
| `docs/` | Setup, deployment, security review, and user guide |

The backend, oracle, contract tests, and deployment commands are Go code in one
module at the repository root (`go.mod`). Hardhat compiles the contract, and the
frontend and the IOTA spike are TypeScript packages with their own
`package.json`. Run the commands below from the repository root.

## Compose development stack

Run `docker compose up -d --build --wait` from the repository root to start all
five services. The `deploy` service creates private-chain accounts, deploys
the registry and records a smoke attestation. Later starts reuse the
configured registry after verifying the saved record. The frontend loads
the public contract address from the deployment file.

The dashboard is at `http://127.0.0.1:5173` and the oracle is at
`http://127.0.0.1:8080`. Set host port overrides in `.env` if another service
uses those ports. To run only infrastructure for local Go and Vite work, use
`docker compose up -d --wait geth verdaccio rustfs`.

Pipeline submissions require the OIDC configuration in the [oracle
guide](../oracle/README.md). Run the IOTA notarization service and Go oracle
on the host using the [IOTA guide](spike-iota.md) when developing IOTA flows.

### Geth shutdown and recovery

Use `docker compose stop` to give Geth time to flush its state. On startup the
deployment step checks the saved smoke record against the chain before the
oracle or dashboard uses the registry. An unclean shutdown can rewind the
head and make the saved deployment unavailable. Keep the old Geth volume
and its deployment receipt together when investigating a rollback.

For a new disposable devnet, the hash-based archive mode stores every state
trie. Select it before creating the volume, and use a separate Compose project
to keep the existing devnet intact. This command starts an isolated Geth and
deploys a registry and smoke record on host port 18545.

```sh
GETH_STATE_SCHEME=hash GETH_GCMODE=archive GETH_HOST_PORT=18545 \
  docker compose -p attestation-archive up -d --build --wait geth deploy
```

Keep the same settings when restarting the archive project. After a forced
stop, Geth can replay its latest transaction in a different block. The
deployment check compares the exact saved block hash and stops on a mismatch. To run
its full stack alongside the default project, also choose free ports for Verdaccio,
RustFS, the oracle, and the frontend. Hash archive mode increases disk use.
See [Geth database shutdown](https://geth.ethereum.org/docs/fundamentals/databases)
and [archive modes](https://geth.ethereum.org/docs/fundamentals/archive).

## Optional IPFS backup for public demo SBOMs

Enable the Kubo profile and list the public demo apps whose SBOMs may be
published. Use exact app names separated by commas.

```sh
IPFS_KUBO_URL=http://kubo:5001 IPFS_PUBLIC_SBOM_APPS=@attestation/demo \
  docker compose --profile ipfs up -d --build --wait
```

The oracle writes the SBOM and scan to RustFS, then tries to pin the SBOM
for an allowlisted app. SQLite stores its CID. Verification tries RustFS
first and uses the CID if the SBOM is unavailable, while still comparing
the recovered bytes to the on-chain hash. Scan reports stay in RustFS.
The Kubo administrative API is reachable only inside the Compose network.

To exercise the fallback with the existing release smoke test, point
`ATTESTATION_SMOKE_IPFS_KUBO_URL` at a disposable Kubo node and use a
published demo package as the smoke fixture.

## Optional Ethereum attestation badges

After deploying the private registry, deploy the separate badge contract
with the configured oracle address as its minter. Load the deployer key
from `ETHEREUM_BADGE_DEPLOYER_PRIVATE_KEY` in your local environment.

```sh
ETHEREUM_BADGE_RPC_URL=http://127.0.0.1:8545 \
  go run ./smart-contract/cmd/deploy-badge --minter "$ETHEREUM_ORACLE_ADDRESS"
```

Set `ETHEREUM_PRIVATE_BADGE_CONTRACT_ADDRESS` to the printed address and
rebuild the Compose stack. For Vite, supply the same address in
`VITE_PRIVATE_BADGE_CONTRACT_ADDRESS`. Sepolia uses `--chain-id 11155111` and
the corresponding Sepolia badge variables from the oracle and frontend
environment examples.

An on-chain verification with result `Valid` queues one badge for the
requester's wallet and record. Badge issuance uses a separate transaction.
The verification result stays available if minting fails, and the oracle
retries pending badge jobs. Connected wallets can view their badges in the
dashboard.

## Contract tooling

Use Go 1.26 or newer for tests and commands:

```sh
cd smart-contract
go test ./registry
```

The tests deploy a fresh registry on go-ethereum's simulated chain for each case,
so they need no running node. `registry/registry_test.go` covers record
integrity, permissions, input limits, replay protection, history pagination, and
the verification request lifecycle. Custom reverts are decoded and checked by
name and arguments.

### Compiling and regenerating bindings

Use Node.js 26 LTS and npm only when the contract changes:

```sh
cd smart-contract
npm ci
npm run build          # compile with Hardhat
go generate ./registry # refresh registry/registry.go from the build artifact
```

`registry/registry.go` is generated and committed, so Go code builds without
Node.js. The Hardhat compiler is pinned to Solidity 0.8.28 with the Cancun EVM
target and optimizer enabled for 200 runs. The contract's version pragma accepts
Solidity from 0.8.28 up to, but excluding, 0.9.0 so editors using newer 0.8.x
compilers can also check the source. Hardhat downloads the compiler on the
first build, which needs internet access and a writable user cache.
`npm run typecheck` checks the Hardhat configuration. `tsx` stays in the
dependencies because the IOTA spike runs through it.

`contracts/AttestationRegistry.sol` implements the Ethereum registry. Its API,
input limits, and events are described in [the contract documentation](attestation-registry.md).
See [private Geth setup](private-geth.md) for persistent deployment. Sepolia and
IOTA testnet are used for the performance comparison. Private-chain
measurements are used for functional validation and the demo.

## Local files

Keep secrets in local environment files or a secret store. `.gitignore`
excludes dependencies, build output, caches, coverage, `.env` and `.env.*`,
SQLite databases and their journal sidecars.
An `.env.example` may be committed when a component defines its configuration.
