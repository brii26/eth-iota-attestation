import assert from "node:assert/strict";
import { randomUUID } from "node:crypto";
import { mkdir, writeFile } from "node:fs/promises";
import { dirname } from "node:path";
import { getContract, parseEther, sha256, toBytes } from "viem";
import { artifact, deploymentFile, privateAccount, privateNetwork, wallet } from "./private-network.js";

async function deploy() {
  const admin = privateAccount("ETHEREUM_ADMIN_PRIVATE_KEY");
  const oracle = privateAccount("ETHEREUM_ORACLE_PRIVATE_KEY");
  if (admin.address === oracle.address) throw new Error("Admin and oracle must use different accounts");
  const { chain, transport, publicClient } = await privateNetwork();
  const adminClient = wallet({ account: admin, chain, transport });
  const oracleClient = wallet({ account: oracle, chain, transport });

  if (process.env.FUND_DEV_ACCOUNTS === "1") {
    const [developer] = await wallet({ chain, transport }).getAddresses();
    if (!developer) throw new Error("No unlocked developer account is available for funding");
    const fundingClient = wallet({ account: developer, chain, transport });
    for (const account of [admin, oracle]) {
      if (await publicClient.getBalance({ address: account.address }) < parseEther("1")) {
        const receipt = await publicClient.waitForTransactionReceipt({
          hash: await fundingClient.sendTransaction({ to: account.address, value: parseEther("10") }),
          timeout: 60_000,
        });
        assert.equal(receipt.status, "success", "Developer account funding failed");
      }
    }
  }

  const deployment = await publicClient.waitForTransactionReceipt({
    hash: await adminClient.deployContract({ abi: artifact.abi, bytecode: artifact.bytecode }),
    timeout: 60_000,
  });
  assert.equal(deployment.status, "success", "Deployment failed");
  assert.ok(deployment.contractAddress, "Deployment did not return a contract address");
  const registry = getContract({
    address: deployment.contractAddress, abi: artifact.abi,
    client: { public: publicClient, wallet: adminClient },
  });
  const permission = await publicClient.waitForTransactionReceipt({
    hash: await registry.write.authorizeOracle([oracle.address]), timeout: 60_000,
  });
  assert.equal(permission.status, "success", "Oracle authorization failed");
  const app = "private-chain-smoke";
  const version = "1.0.0";
  const scanId = randomUUID();
  const hashes = {
    artifactHash: sha256(toBytes("demo artifact bytes")),
    sbomHash: sha256(toBytes("demo SBOM bytes")),
    scanHash: sha256(toBytes("demo scan bytes")),
  };
  const recordReceipt = await publicClient.waitForTransactionReceipt({
    hash: await registry.write.addRecord(
      [app, version, scanId, hashes.artifactHash, hashes.sbomHash, hashes.scanHash],
      { account: oracleClient.account },
    ),
    timeout: 60_000,
  });
  assert.equal(recordReceipt.status, "success", "Smoke record transaction failed");
  const latest = await registry.read.getLatestRecord([app, version]);
  assert.equal(latest.scanId, scanId);
  assert.equal(latest.signer.toLowerCase(), oracle.address.toLowerCase());
  const recordId = await registry.read.recordCount();
  const result = {
    chainId: chain.id,
    contractAddress: registry.address,
    admin: admin.address,
    oracle: oracle.address,
    deploymentTx: deployment.transactionHash,
    permissionTx: permission.transactionHash,
    recordTx: recordReceipt.transactionHash,
    recordBlockHash: recordReceipt.blockHash,
    recordId: recordId.toString(),
    app, version, scanId, ...hashes,
    timestamp: latest.timestamp.toString(),
  };
  await mkdir(dirname(deploymentFile), { recursive: true });
  await writeFile(deploymentFile, `${JSON.stringify(result, null, 2)}\n`, { mode: 0o600 });
  console.log(JSON.stringify(result, null, 2));
}

await deploy();
