import { createHash, randomUUID } from "node:crypto";
import { readFile } from "node:fs/promises";

const [manifestPath, archivePath, sbomPath, scanPath] = process.argv.slice(2);
if (!scanPath) throw new Error("Usage: submit-attestation.mjs package.json artifact.tgz sbom.json scan.json");
const configuration = Object.fromEntries([
  "ORACLE_URL", "ATTESTATION_NETWORK", "OIDC_AUDIENCE",
  "ACTIONS_ID_TOKEN_REQUEST_URL", "ACTIONS_ID_TOKEN_REQUEST_TOKEN",
].map((key) => {
  const value = process.env[key];
  if (!value) throw new Error(`${key} is required`);
  return [key, value];
}));

const [manifest, archive, sbom, scan] = await Promise.all(
  [manifestPath, archivePath, sbomPath, scanPath].map((path) => readFile(path)),
);
const { name: app, version } = JSON.parse(manifest);
const hash = (data) => createHash("sha256").update(data).digest("hex");
const tokenURL = new URL(configuration.ACTIONS_ID_TOKEN_REQUEST_URL);
tokenURL.searchParams.set("audience", configuration.OIDC_AUDIENCE);
const tokenResponse = await fetch(tokenURL, {
  headers: { Authorization: `Bearer ${configuration.ACTIONS_ID_TOKEN_REQUEST_TOKEN}` },
  signal: AbortSignal.timeout(15_000),
});
if (!tokenResponse.ok) throw new Error(`GitHub OIDC returned HTTP ${tokenResponse.status}`);
const { value: token } = await tokenResponse.json();
if (typeof token !== "string" || !token) throw new Error("GitHub OIDC did not return a token");

const network = configuration.ATTESTATION_NETWORK;
const scanId = randomUUID();
const endpoint = new URL("/api/v1/attestations", configuration.ORACLE_URL);
const response = await fetch(endpoint, {
  method: "POST",
  headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  body: JSON.stringify({
    network, app, version, scanId,
    claimedHashes: { artifact: hash(archive), sbom: hash(sbom), scan: hash(scan) },
    sbomBase64: sbom.toString("base64"),
    scanBase64: scan.toString("base64"),
  }),
  signal: AbortSignal.timeout(120_000),
});
const result = await response.json();
if (!response.ok) throw new Error(`Oracle rejected ${app}@${version}: ${result.error?.code || response.status}`);
if (result.status === "included") {
  console.log(JSON.stringify(result));
} else if (result.status === "pending") {
  const statusURL = new URL(`${endpoint.pathname}/${encodeURIComponent(network)}/${encodeURIComponent(scanId)}`, endpoint);
  for (let attempt = 0; attempt < 36; attempt++) {
    await new Promise((resolve) => setTimeout(resolve, 5_000));
    const response = await fetch(statusURL, { signal: AbortSignal.timeout(15_000) });
    if (!response.ok) throw new Error(`Oracle status returned HTTP ${response.status}`);
    const status = await response.json();
    if (status.status === "included") {
      console.log(JSON.stringify(status));
      break;
    }
    if (status.status === "failed" || attempt === 35) throw new Error(`Attestation ${scanId} is ${status.status}`);
  }
} else {
  throw new Error(`Oracle returned unexpected status for scan ${scanId}`);
}
