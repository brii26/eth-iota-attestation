# Benchmark dataset

The generator creates ten synthetic npm packages with 2, 24, or 96
bundled dependencies. It packages each release, scans the same archive
with Syft, then scans the resulting CycloneDX SBOM with Grype. Each run gets
a new prerelease version. The JSON manifest records file sizes, SHA-256
fingerprints, and tool versions for the shared RQ1 and RQ2 dataset.

Install Node.js 26, Syft 1.57.2, and Grype 0.122.1. Run from the repository root.

```sh
go run ./oracle/test/fixtures -output /tmp/attestation-datasets
```

To publish the dataset, add `-registry http://127.0.0.1:4873` and
`-npm-userconfig /path/to/authorized.npmrc`. Give the same generated
`manifest.json` to both benchmark runners. File paths in the manifest
are relative to its directory.
