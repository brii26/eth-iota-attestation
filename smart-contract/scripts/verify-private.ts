import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { getContract, isAddress, isHex } from "viem";
import { artifact, deploymentFile, privateNetwork } from "./private-network.js";

const saved: Record<string, unknown> = JSON.parse(await readFile(deploymentFile, "utf8"));
assert.equal(saved.chainId, 1337, "Deployment receipt must belong to chain ID 1337");
assert.ok(typeof saved.contractAddress === "string" && isAddress(saved.contractAddress));
assert.ok(typeof saved.recordId === "string" && /^[1-9][0-9]*$/.test(saved.recordId));
assert.ok(typeof saved.recordTx === "string" && isHex(saved.recordTx));
const { publicClient } = await privateNetwork();
const registry = getContract({ address: saved.contractAddress, abi: artifact.abi, client: publicClient });
const receipt = await publicClient.getTransactionReceipt({ hash: saved.recordTx });
assert.equal(receipt.status, "success");
assert.equal(receipt.blockHash, saved.recordBlockHash, "Recorded transaction changed chains or blocks");
const record = await registry.read.getRecord([BigInt(saved.recordId)]);
for (const field of ["app", "version", "scanId", "artifactHash", "sbomHash", "scanHash"] as const) {
  assert.equal(record[field], saved[field], `Stored ${field} differs from deployment receipt`);
}
assert.equal(record.signer.toLowerCase(), String(saved.oracle).toLowerCase());
assert.equal(record.timestamp.toString(), saved.timestamp);
assert.equal((await registry.read.admin()).toLowerCase(), String(saved.admin).toLowerCase());
console.log(JSON.stringify({
  status: "verified", chainId: 1337, contractAddress: registry.address,
  recordId: saved.recordId, scanId: record.scanId, blockHash: receipt.blockHash,
}, null, 2));
