# Development setup

## Repository layout

| Path | Responsibility |
| --- | --- |
| `frontend/` | React dashboard and MetaMask integration |
| `smart-contract/` | Hardhat 3, Solidity contracts, Node test runner, and viem |
| `oracle/` | Fastify backend, oracle, chain adapters, and SQLite index |
| `docs/` | Setup, deployment, security review, and user guide |

Each application component manages its own dependencies. Run the commands
below from the repository root.

## Contract tooling

Use Node.js 24 LTS and npm. Install dependencies from the committed lockfile:

```sh
cd smart-contract
npm ci
npm run build
npm test
npm run typecheck
```

`npx hardhat build` is equivalent to `npm run build`. The configuration loads
`@nomicfoundation/hardhat-toolbox-viem`, including the Node test runner and viem.
Solidity is pinned to 0.8.28 with the Cancun EVM target and optimizer enabled
for 200 runs. Hardhat downloads the compiler on the first build, which needs
internet access and a writable user cache.

This scaffold has no application contracts or tests yet. `contracts/` is reserved
for Solidity contracts and `test/` for their tests. A successful empty build or
test run validates tooling only.

The configured `hardhatMainnet` network is a local simulated Ethereum L1 for
development. Private Geth deployment is not configured yet. Sepolia and IOTA
testnet are used for the performance comparison; private-chain measurements
are used for functional validation and the demo.

## Local files

Keep secrets in local environment files or a secret store. `.gitignore`
excludes dependencies, build output, caches, coverage, `.env` and `.env.*`,
SQLite databases and their journal sidecars.
An `.env.example` may be committed when a component defines its configuration.
