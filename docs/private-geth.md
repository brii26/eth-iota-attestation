# Private Geth development chain

The Compose service runs a single Geth developer node with no peers. A named
volume preserves its chain and developer account. This network supports
functional development and the demonstration, separate from testnet benchmarks.

## Start the node

Install Docker Engine with Compose and Node.js 24 LTS, then run from the
repository root:

```sh
docker compose up -d --wait geth
curl -fsS http://127.0.0.1:8545 \
  -H 'Content-Type: application/json' \
  --data '{"jsonrpc":"2.0","id":1,"method":"eth_chainId","params":[]}'
```

The expected chain ID is `0x539` (1337). The pinned image is
`ethereum/client-go:v1.17.10`. RPC is bound to the host loopback interface.
The developer account is unlocked and funded by the development genesis.
Keep this RPC local to the development machine.

## Create application accounts

From `smart-contract/`, install dependencies and create an environment file:

```sh
npm ci
cp .env.example .env
chmod 600 .env
```

Set `ETHEREUM_ADMIN_PRIVATE_KEY` and `ETHEREUM_ORACLE_PRIVATE_KEY` to two
different development account keys in `.env`. Keys must be 32-byte hex strings
starting with `0x`. These files are ignored by Git. Use development accounts
that do not control assets on other networks.

Set `FUND_DEV_ACCOUNTS=1` for the first deployment. This explicitly enables
funding from the unlocked Geth developer account. The script refuses any RPC
with a chain ID other than 1337. Set the flag back to `0` after funding.

## Deploy and write a smoke record

```sh
npm run build
npm run geth:deploy
```

The script signs deployment and oracle authorization with the admin account,
then signs one `addRecord` transaction with the oracle account. It reads back
the latest record and prints JSON containing the contract address, transaction
hashes, and record metadata. Only fingerprints of fixed demo strings are used
for this smoke record, without invoking the production scanning pipeline.

The same public metadata is saved to `deployments/ethereum-private.json`.
Each deployment creates a new registry and replaces that file. Preserve it
when testing a particular deployment. It contains no private keys.

## Check persistence

From the repository root, restart the service without deleting its volume:

```sh
docker compose restart geth
docker compose up -d --wait geth
cd smart-contract
npm run geth:verify
```

Verification uses the saved deployment receipt and requires no signing key.
It checks that the transaction still exists in the same block and that the
record's fields, signer, timestamp, and administrator match. Recreating the
container with the same named volume also retains the chain.
`docker compose down -v` deletes the chain and invalidates saved receipts.

## Connect MetaMask

Add a custom network with RPC URL `http://127.0.0.1:8545`, chain ID `1337`,
currency symbol `ETH`, and any descriptive name. Import the development admin
account if administration transactions are needed. A successful connection
should display the balance funded during deployment. For a node on another
machine, forward its loopback RPC over SSH and use the forwarded local port.

## Validation status

Deployment, developer-account funding, separate admin and oracle signatures,
record read-back, chain ID validation, and tampered-receipt rejection have been
checked on a simulated Ethereum chain through an in-process RPC transport.
Actual Geth startup, volume persistence, and MetaMask connection remain
unverified and must be exercised on a host with Docker access. The simulation
does not provide evidence of Geth-specific persistence or wallet behavior.

The course's acceptance of single-node `--dev` as the required private chain
remains a deployment decision to confirm with the teaching team.

Sources: [Geth developer mode](https://geth.ethereum.org/docs/developers/dapp-developer/dev-mode)
and [Geth v1.17.10](https://github.com/ethereum/go-ethereum/releases/tag/v1.17.10).
