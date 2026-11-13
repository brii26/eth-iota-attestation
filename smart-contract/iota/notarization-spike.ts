import assert from "node:assert/strict";
import { createHash, randomUUID } from "node:crypto";
import { mkdir, writeFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import { setTimeout as sleep } from "node:timers/promises";

const runId = randomUUID();
const encode = (value: unknown) => JSON.stringify(value, (_, item) =>
  typeof item === "bigint" ? item.toString() : item, 2);
const fingerprint = (value: string) => `0x${createHash("sha256").update(value).digest("hex")}`;
const payload = (scanId: string) => JSON.stringify({
  schemaVersion: 1,
  app: "iota-spike",
  version: "1.0.0",
  scanId,
  artifactHash: fingerprint("demo artifact bytes"),
  sbomHash: fingerprint("demo SBOM bytes"),
  scanHash: fingerprint("demo scan bytes"),
});
const report: Record<string, unknown> = {
  runId,
  recordedAt: new Date().toISOString(),
  sdkVersions: { notarization: "0.1.15", iotaSdk: "1.17.0", interaction: "0.14.0" },
  status: "running",
};
type Observation = {
  name: string;
  status: "passed" | "failed" | "skipped";
  evidence?: unknown;
};
const observations: Observation[] = [];
report.observations = observations;
const createdObjects: string[] = [];
report.createdObjects = createdObjects;
const outputFile = fileURLToPath(new URL(`results/${runId}.json`, import.meta.url));
async function save() {
  await mkdir(fileURLToPath(new URL("results/", import.meta.url)), { recursive: true });
  await writeFile(outputFile, `${encode(report)}\n`, { mode: 0o600 });
}

function message(error: unknown) {
  let text = error instanceof Error ? error.message : String(error);
  const secret = process.env.IOTA_PRIVATE_KEY;
  if (secret) text = text.replaceAll(secret, "[redacted]");
  return text;
}

async function main() {
  if (process.argv.includes("--check")) {
    const decoded = JSON.parse(payload(runId));
    for (const field of ["artifactHash", "sbomHash", "scanHash"]) {
      assert.match(decoded[field], /^0x[0-9a-f]{64}$/);
    }
    assert.equal(decoded.scanId, runId);
    report.status = "offline-check-only";
    report.payloadExample = decoded;
    report.note = "No SDK was loaded and no on-chain scenario was executed";
    return;
  }

  const network = process.env.IOTA_NETWORK ?? "testnet";
  if (network !== "testnet" && network !== "localnet") {
    throw new Error("This spike only supports testnet or localnet");
  }
  let packageId = process.env.IOTA_NOTARIZATION_PKG_ID;
  const lockSeconds = Number(process.env.IOTA_LOCK_SECONDS ?? "120");
  if (!Number.isSafeInteger(lockSeconds) || lockSeconds < 30 || lockSeconds > 600) {
    throw new Error("IOTA_LOCK_SECONDS must be an integer from 30 to 600");
  }
  const [notarization, sdk, accounts, faucet, interaction] = await Promise.all([
    import("@iota/notarization/node"),
    import("@iota/iota-sdk/client"),
    import("@iota/iota-sdk/keypairs/ed25519"),
    import("@iota/iota-sdk/faucet"),
    import("@iota/iota-interaction-ts/node"),
  ]);
  const { NotarizationClient, NotarizationClientReadOnly, TimeLock, State } = notarization;
  const rpcUrl = process.env.IOTA_RPC_URL ?? sdk.getFullnodeUrl(network);
  const rpc = new sdk.IotaClient({ url: rpcUrl });

  let readOnly: any;
  if (packageId && /^0x[0-9a-fA-F]{64}$/.test(packageId)) {
    readOnly = await NotarizationClientReadOnly.createWithPkgId(rpc as any, packageId);
  } else {
    try {
      readOnly = await NotarizationClientReadOnly.create(rpc as any);
      packageId = readOnly.packageId();
    } catch {
      throw new Error("Set IOTA_NOTARIZATION_PKG_ID to the deployed notarization package ID");
    }
  }

  const secret = process.env.IOTA_PRIVATE_KEY;
  let keypair;
  if (!secret) keypair = accounts.Ed25519Keypair.generate();
  else if (/^(0x)?[0-9a-fA-F]{64}$/.test(secret)) {
    keypair = accounts.Ed25519Keypair.fromSecretKey(Uint8Array.from(Buffer.from(secret.replace(/^0x/, ""), "hex")));
  } else {
    try { keypair = accounts.Ed25519Keypair.fromSecretKey(secret); }
    catch { throw new Error("IOTA_PRIVATE_KEY is not a supported Ed25519 private key"); }
  }
  const address = keypair.toIotaAddress();
  report.network = network;
  report.rpcOrigin = new URL(rpcUrl).origin;
  report.packageId = packageId;
  report.sender = address;
  report.ephemeralAccount = !secret;
  const chainIdentifier = await rpc.getChainIdentifier();
  report.chainIdentifier = chainIdentifier;
  if (network === "testnet") {
    const officialUrl = sdk.getFullnodeUrl("testnet");
    const officialIdentifier = rpcUrl === officialUrl ? chainIdentifier
      : await new sdk.IotaClient({ url: officialUrl }).getChainIdentifier();
    assert.equal(chainIdentifier, officialIdentifier, "RPC does not match the official testnet chain");
  }
  if (process.env.IOTA_REQUEST_FAUCET === "1") {
    try {
      const host = process.env.IOTA_FAUCET_URL ?? faucet.getFaucetHost(network);
      await faucet.requestIotaFromFaucetV0({ host, recipient: address });
    } catch (err) {
      console.warn(`Faucet request skipped or failed: ${message(err)}`);
    }
  }
  if (BigInt((await rpc.getBalance({ owner: address })).totalBalance) === 0n) {
    throw new Error("Fund the spike account or enable IOTA_REQUEST_FAUCET before running");
  }
  const client = await NotarizationClient.create(readOnly, new interaction.test_utils.Ed25519KeypairSigner(keypair as any));

  async function ledgerTimeMs() {
    const latest = await rpc.getLatestCheckpointSequenceNumber();
    return BigInt((await rpc.getCheckpoint({ id: latest })).timestampMs);
  }

  async function transactionEvidence(digest: string) {
    const response = await rpc.waitForTransaction({
      digest, options: { showEffects: true, showInput: true, showObjectChanges: true, showEvents: true },
    });
    const gas = response.effects?.gasUsed;
    return {
      digest: response.digest,
      status: response.effects?.status,
      sender: response.transaction?.data.sender,
      checkpoint: response.checkpoint,
      timestampMs: response.timestampMs,
      gasUsed: gas,
      netGasCost: gas ? (BigInt(gas.computationCost) + BigInt(gas.storageCost) - BigInt(gas.storageRebate)).toString() : null,
      objectChanges: response.objectChanges,
      events: response.events,
    };
  }

  async function create(scanId: string, unlockAt?: number) {
    let builder = client.createLocked()
      .withStringState(payload(scanId), "artifact-attestation-v1")
      .withImmutableDescription("Software artifact attestation spike");
    if (unlockAt !== undefined) builder = builder.withDeleteLock(TimeLock.withUnlockAt(unlockAt));
    const { output, response } = await builder.finish().buildAndExecute(client);
    const evidence = await transactionEvidence(response.digest);
    assert.equal(evidence.status?.status, "success", "Create transaction failed");
    assert.equal(evidence.sender?.toLowerCase(), address.toLowerCase());
    createdObjects.push(output.id);
    observations.push({ name: `create:${scanId}`, status: "passed", evidence: { objectId: output.id, ...evidence } });
    await save();
    return output;
  }

  async function expectRejected(name: string, operation: () => Promise<unknown>) {
    let confirmed;
    try {
      const result = await operation() as { response?: { digest?: string } };
      if (result.response?.digest) {
        confirmed = await transactionEvidence(result.response.digest);
        if (confirmed.status?.status === "failure") throw new Error(confirmed.status.error);
      }
    } catch (error) {
      const reason = message(error);
      // RPC outages, gas exhaustion, and signature errors are not lock evidence.
      const moveAbort = /MoveAbort/i.test(reason) && /notarization|time.?lock|locking/i.test(reason);
      const invalidLock = /delete.?lock/i.test(reason) && /UntilDestroyed|invalid|not allowed/i.test(reason);
      if (!moveAbort && !invalidLock) {
        observations.push({ name, status: "failed", evidence: { reason, kind: "unclassified-error" } });
        throw new Error(`${name} failed without evidence of a lock rejection`);
      }
      observations.push({ name, status: "passed", evidence: {
        reason, transaction: confirmed, kind: moveAbort ? "move-abort" : "sdk-validation",
      } });
      await save();
      return;
    }
    observations.push({ name, status: "failed", evidence: "Operation unexpectedly succeeded" });
    throw new Error(`${name} unexpectedly succeeded`);
  }

  const unlockAt = Number(await ledgerTimeMs() / 1000n) + lockSeconds;
  report.unlockAtSeconds = unlockAt;
  const locked = await create(`${runId}:locked`, unlockAt);
  const fetched = await readOnly.getNotarizationById(locked.id);
  assert.equal(fetched.method, "Locked");
  assert.equal(fetched.state.data.toString(), payload(`${runId}:locked`));
  assert.equal(fetched.owner.toLowerCase(), address.toLowerCase());
  const createdAt = BigInt(await readOnly.createdAtTs(locked.id));
  const nowMs = await ledgerTimeMs();
  assert.ok(createdAt <= nowMs && nowMs - createdAt < 300_000n, "Creation timestamp must match the ledger millisecond clock");
  assert.equal(await readOnly.isUpdateLocked(locked.id), true);
  assert.equal(await readOnly.isTransferLocked(locked.id), true);
  observations.push({ name: "read-back", status: "passed", evidence: {
    objectId: locked.id, owner: fetched.owner, createdAtMs: createdAt.toString(),
    locking: fetched.immutableMetadata.locking,
  } });
  await expectRejected("update-locked-state", async () =>
    client.updateState(State.fromString("changed", "tampered"), locked.id).buildAndExecute(client));
  assert.equal((await readOnly.state(locked.id)).data.toString(), payload(`${runId}:locked`));
  assert.ok(await ledgerTimeMs() < BigInt(unlockAt) * 1000n, "Delete lock expired before the rejection probe");
  await expectRejected("destroy-before-unlock", async () => client.destroy(locked.id).buildAndExecute(client));
  const permanentUpdateLock = fetched.immutableMetadata.locking?.updateLock;
  assert.ok(permanentUpdateLock && permanentUpdateLock.type === "UntilDestroyed");
  await expectRejected("until-destroyed-delete-lock", async () =>
    client.createLocked().withStringState(payload(`${runId}:invalid-lock`))
      .withDeleteLock(permanentUpdateLock).finish().buildAndExecute(client));

  const unlocked = await create(`${runId}:no-delete-lock`);
  assert.equal(await readOnly.isDestroyAllowed(unlocked.id), true);
  const destroyed = await client.destroy(unlocked.id).buildAndExecute(client);
  const destroyedEvidence = await transactionEvidence(destroyed.response.digest);
  assert.equal(destroyedEvidence.status?.status, "success");
  observations.push({ name: "destroy-without-delete-lock", status: "passed", evidence: destroyedEvidence });

  if (process.env.IOTA_WAIT_FOR_UNLOCK !== "0") {
    const deadline = Date.now() + (lockSeconds + 60) * 1000;
    while (await ledgerTimeMs() < BigInt(unlockAt) * 1000n) {
      if (Date.now() > deadline) throw new Error("Timed out waiting for the ledger delete-lock expiry");
      await sleep(1000);
    }
    assert.equal(await readOnly.isDestroyAllowed(locked.id), true);
    const result = await client.destroy(locked.id).buildAndExecute(client);
    const evidence = await transactionEvidence(result.response.digest);
    assert.equal(evidence.status?.status, "success");
    observations.push({ name: "destroy-after-unlock", status: "passed", evidence });
  } else {
    observations.push({ name: "destroy-after-unlock", status: "skipped", evidence: "IOTA_WAIT_FOR_UNLOCK=0" });
  }

  if (process.env.IOTA_TEST_PARALLEL === "1") {
    // Run last because conflicting gas objects may remain locked for the epoch.
    const attempts = await Promise.allSettled([create(`${runId}:parallel-1`), create(`${runId}:parallel-2`)]);
    const outcomes = attempts.map((attempt) => {
      if (attempt.status === "fulfilled") return { status: "fulfilled", objectId: attempt.value.id };
      const reason = message(attempt.reason);
      const gasConflict = /equivocation|object.*(?:locked|version.*unavailable)|gas.*(?:conflict|locked)/i.test(reason);
      return { status: gasConflict ? "gas-object-contention" : "unclassified-error", reason };
    });
    observations.push({ name: "parallel-account-writes",
      status: outcomes.some((item) => item.status === "unclassified-error") ? "failed" : "passed",
      evidence: outcomes,
    });
  } else {
    observations.push({ name: "parallel-account-writes", status: "skipped", evidence: "IOTA_TEST_PARALLEL is disabled" });
  }

  const ownedIds: string[] = [];
  let cursor: string | null | undefined;
  let truncated = false;
  for (let page = 0; page < 20; page++) {
    const objects = await rpc.getOwnedObjects({ owner: address, cursor, limit: 50 });
    ownedIds.push(...objects.data.flatMap((object: { data?: { objectId: string } | null }) =>
      object.data ? [object.data.objectId] : []));
    if (!objects.hasNextPage) break;
    cursor = objects.nextCursor;
    if (page === 19) truncated = true;
  }
  observations.push({ name: "owned-object-list", status: "passed", evidence: { owner: address, objectIds: ownedIds, truncated } });
  report.status = observations.some((item) => item.status !== "passed") ? "incomplete" : "completed";
  if (report.status !== "completed") process.exitCode = 2;
}

try {
  await main();
} catch (error) {
  report.status = "failed";
  report.error = message(error);
  process.exitCode = 1;
} finally {
  await save();
  console.log(encode(report));
  console.log(`Evidence file: ${outputFile}`);
}
