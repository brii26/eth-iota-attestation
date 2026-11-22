# RQ1 integrity matrix

Generate the shared [fixture dataset](../oracle/test/fixtures/README.md)
with at least ten published package and version variations. Use a dedicated
private Geth registry, a deployed [three-hash baseline](baseline.md), and a
separate funded IOTA account. Configure the development networks and
storage with the variables in [the oracle configuration](../oracle/.env.example).
Set `BASELINE_ETHEREUM_CONTRACT` to the baseline registry address.

Save the filename of the published manifest and the desired JSON output in
a local configuration file.

```json
{
  "manifest": "/path/to/manifest.json",
  "output": "/path/to/rq1-results.json"
}
```

```sh
ATTESTATION_RQ1_CONFIG=/path/to/rq1.json \
  go test -count=1 -timeout 3h -run '^TestRQ1IntegrityMatrix$' -v ./oracle/internal/benchmark
```

The runner records fixtures on the two configured ledgers. It checks
individual file changes, wrong versions, stale scans, duplicate scan IDs,
and unauthorized pipeline requests. File changes are injected into the
test verifier's read path. Each case reports its observed outcome and the
component enforcing the rule. The runner records mismatches in the JSON
before returning a failing test.

The stale-index check changes the order in a temporary SQLite database
and compares ledger discovery against a known older and newer record.
The signer scenario exercises verifier classification with a deny policy.
Live oracle permission changes and direct IOTA duplicate writes are
separate scenarios still tracked in issue #17.
