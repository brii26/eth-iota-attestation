# RQ2 testnet benchmark

Use one machine for the Sepolia and IOTA testnet runs. Deploy the registry
and fund the testnet oracle accounts before starting. Start the IOTA
notarization service and configure the same signing account and delete-lock
policy for the whole run. Generate and publish the shared
[fixture dataset](../oracle/test/fixtures/README.md).

The runner uses the standard oracle API in-process. It reads the package
from Verdaccio, stores the reports in RustFS, and submits real transactions.
Use separate benchmark storage and SQLite data. Supply the Sepolia, RustFS,
and IOTA environment variables from [the oracle configuration](../oracle/.env.example).
The test harness authorizes only the packages in its manifest.

Create a local configuration file.

```json
{
  "manifest": "/path/to/manifest.json",
  "output": "/path/to/rq2-results.json",
  "index": "/path/to/rq2.sqlite",
  "bursts": [1, 4, 8],
  "repetitions": 2
}
```

Scale the burst sizes and repetitions to the available balances before
starting. Run the same workload across different time windows and save
a separate output file for each.

```sh
ATTESTATION_RQ2_CONFIG=/path/to/rq2.json \
  go test -count=1 -timeout 3h -run '^TestRQ2TestnetComparison$' -v ./oracle/internal/benchmark
```

The runner alternates network order between rounds and follows finality
in parallel. It measures observed submit-to-inclusion and finality times
on both networks. Ethereum finality requires the receipt's block to be
covered by the finalized head and remain canonical. IOTA finality requires
a checkpoint for the transaction. The JSON includes sample failures,
inclusion throughput, oracle processing, local hashing, chain reads,
verification time, and summary statistics.

Cost uses native units. Ethereum reports gas used and fees in wei. IOTA
reports computation cost, storage cost, and storage rebate. Payload bytes
count Ethereum calldata or IOTA notarization state. Transaction bytes use
each chain SDK's serialized transaction size.
