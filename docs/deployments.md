# Network deployments

## Private Ethereum

Use the [private Geth guide](private-geth.md) for local deployment and the
functional demo. Keep private-chain results separate from testnet measurements.

## Sepolia

After the registry version is frozen, configure `ETHEREUM_SEPOLIA_RPC_URL`,
`ETHEREUM_ADMIN_PRIVATE_KEY`, and `ETHEREUM_ORACLE_PRIVATE_KEY` in the
environment. Use separate funded Sepolia accounts. The command validates chain
ID 11155111, deploys the registry, authorizes the oracle, and waits for its smoke
record to finalize.

```sh
go run ./smart-contract/cmd/deploy-sepolia
```

Set `ETHEREUM_SEPOLIA_CONTRACT_ADDRESS` to the resulting contract address.
Configure `ETHEREUM_SEPOLIA_BACKUP_RPC_URL` for oracle failover. Keep the
public deployment JSON together with the Git commit and the Solidity version
used to build the bindings. Record the issued deployment address here after
the testnet run.

Budget the oracle balance using the measured gas per write, the number of
planned writes, and the expected fee cap. Include deployment and authorization
in the separate admin balance. The deployment output includes the smoke write
gas used so the team can refine the estimate before running the benchmark.

## IOTA testnet

Use a funded oracle account with the settings in the [IOTA guide](spike-iota.md).
Budget for both computation and storage. Record the notarization package ID,
checkpoint, and delete-lock configuration with the benchmark results.


## Pipeline access over Tailscale

Install Tailscale on the target host and add that machine to the tailnet.
Set `PIPELINE_BIND_ADDRESS` in the repository root environment file to its
Tailscale IP. Compose publishes the Verdaccio and oracle ports on that
address. Use the same address in the GitHub repository's `VERDACCIO_URL`
and `ORACLE_URL` variables. If the verifier runs on the target host, the
frontend still uses the private Docker network for API calls.

Tag the runner with `tag:ci` and scope its tailnet access to that host's
Verdaccio and oracle TCP ports, 4873 and 8080 by default. Configure the
GitHub OIDC settings using the [oracle guide](../oracle/README.md). Run
the workflow and confirm it reaches both services and writes a record.
