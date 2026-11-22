# Registry gas and static analysis

Solidity 0.8.28 compiled for Cancun with 200 optimizer runs. The Go test uses
the same simulated ledger and inputs for the five transaction receipts.
`addRecord` now checks calldata lengths directly and copies the validated
fields into storage without a temporary memory record. The ABI and record
layout are identical.

| Method | Before | After | Gas change |
| --- | ---: | ---: | ---: |
| `addRecord` | 1,357,127 | 1,356,091 | −1,036 |
| `authorizeOracle` | 133,757 | 133,757 | 0 |
| `revokeOracle` | 28,624 | 28,624 | 0 |
| `requestVerification` | 466,155 | 466,155 | 0 |
| `fulfillVerification` | 472,552 | 472,552 | 0 |

These measurements come from the test ledger. Benchmark testnet execution
costs separately using the deployed registry and its actual receipts.

```sh
go test -count=1 -run TestRegistryGasReport -v ./smart-contract/registry
go test ./smart-contract/registry
```

Slither 0.11.6 ran 102 detectors on `AttestationRegistry.sol` and reported
zero findings. To reproduce, use the same compiler settings as Hardhat.

```sh
slither smart-contract/contracts/AttestationRegistry.sol \
  --solc /path/to/solc-0.8.28 \
  --solc-args '--optimize --optimize-runs 200 --evm-version cancun'
```
