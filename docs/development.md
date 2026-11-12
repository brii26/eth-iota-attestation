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
