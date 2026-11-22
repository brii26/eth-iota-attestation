# Three-hash baseline

The RQ1 baseline records and checks the package, SBOM, and scan SHA-256
fingerprints. The same files still match if the release version or scan ID
changes, because neither identity is part of the baseline record.

Set `BASELINE_ETHEREUM_RPC_URL` and `BASELINE_ETHEREUM_PRIVATE_KEY` for the
selected development network. Set `BASELINE_ETHEREUM_WRITER` to the oracle
address and deploy from the repository root.

```sh
go run ./oracle/cmd/baseline -network ethereum-private -action deploy
```

Set `BASELINE_ETHEREUM_CONTRACT` to the address printed by deployment. Use the
writer key for `-action write` and pass `-artifact`, `-sbom`, and `-scan`
with the three hexadecimal fingerprints. Run `-action verify` with the same
fingerprints to query whether the contract recorded them. Verification does
not need a signing key. Sepolia uses `-network ethereum-sepolia`.

For IOTA, run the notarization service with the same funded oracle account as
the full attestation test. Set `IOTA_SERVICE_URL`, `IOTA_SERVICE_TOKEN`, and
`IOTA_TRUSTED_SIGNERS`. Write using `-network iota-testnet -action write`
and the same fingerprint flags. Pass the returned object ID as `-record`
when verifying. The service accepts only locked baseline objects and checks
the creation transaction's signer.
