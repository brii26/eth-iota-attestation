# Artifact attestation dashboard

Install Node.js 26, run `npm ci` and `npm run dev` from this directory, then
open the URL printed by Vite. The dev server proxies requests to the Go
oracle at `http://127.0.0.1:8080` by default. Use `ORACLE_URL` to change the
proxy destination. The oracle supplies the list of enabled networks.

Copy `.env.example` to `.env` and set `VITE_PRIVATE_REGISTRY` and
`VITE_SEPOLIA_REGISTRY` to the matching deployed addresses for direct Ethereum
reads. Use `VITE_PRIVATE_RPC_URL` or `VITE_SEPOLIA_RPC_URL` to select the public
read endpoint. Vite exposes those values to browsers, so use public RPC URLs.

Find a release by network, package name, and version. Leave the latest record
selected or choose an included scan in history, then verify. The dashboard
shows the integrity checks, scan freshness, and the independent Ethereum read.

`npm run build` creates `dist/`. For deployment, serve that directory and
forward `/api/` on the same origin to the oracle.

## MetaMask

Select an Ethereum network and click **Connect MetaMask**. The dashboard asks
the wallet to switch to the expected chain and offers the private Geth network
configuration when necessary. The registry administrator can authorize or
revoke an oracle by address. Any account on the selected network can request
verification of a selected record or the latest verified record. The
dashboard reads the request ID from the mined transaction and polls for the
oracle's result.
