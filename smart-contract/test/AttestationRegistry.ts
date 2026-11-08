import assert from "node:assert/strict";
import { after, describe, it } from "node:test";
import { network } from "hardhat";
import { encodeAbiParameters, keccak256, maxUint256, zeroAddress, zeroHash } from "viem";
import type { Hex } from "viem";
import type {} from "../artifacts/contracts/AttestationRegistry.sol/artifacts.d.ts";

const hashes = ["11", "22", "33"].map((byte) => `0x${byte.repeat(32)}` as Hex);
const [artifactHash, sbomHash, scanHash] = hashes as [Hex, Hex, Hex];
const record = (scanId = "scan-1", app = "demo-lib", version = "1.0.0") =>
  [app, version, scanId, artifactHash, sbomHash, scanHash] as const;

describe("AttestationRegistry", async () => {
  const connection = await network.create();
  const { viem, networkHelpers } = connection;
  const publicClient = await viem.getPublicClient();
  after(async () => connection.close());

  async function deployFixture() {
    const [admin, oracle, stranger] = await viem.getWalletClients();
    const registry = await viem.deployContract("AttestationRegistry");
    const asOracle = { account: oracle.account };
    const asStranger = { account: stranger.account };
    await publicClient.waitForTransactionReceipt({
      hash: await registry.write.authorizeOracle([oracle.account.address]),
    });
    return { registry, admin, oracle, stranger, asOracle, asStranger };
  }

  async function recordFixture() {
    const fixture = await deployFixture();
    const receipt = await publicClient.waitForTransactionReceipt({
      hash: await fixture.registry.write.addRecord(record(), fixture.asOracle),
    });
    return { ...fixture, receipt };
  }

  async function requestFixture() {
    const fixture = await recordFixture();
    const requestReceipt = await publicClient.waitForTransactionReceipt({
      hash: await fixture.registry.write.requestVerification([1n], fixture.asStranger),
    });
    return { ...fixture, requestReceipt };
  }

  it("sets the deployer as admin without granting automatic oracle access", async () => {
    const { registry, admin, oracle } = await networkHelpers.loadFixture(deployFixture);
    assert.equal((await registry.read.admin()).toLowerCase(), admin.account.address);
    assert.equal(await registry.read.isAuthorizedOracle([admin.account.address]), false);
    assert.equal(await registry.read.isAuthorizedOracle([oracle.account.address]), true);
  });

  it("stores exact hashes with the transaction signer and block timestamp", async () => {
    const { registry, oracle, receipt } = await networkHelpers.loadFixture(recordFixture);
    const stored = await registry.read.getRecord([1n]);
    const block = await publicClient.getBlock({ blockHash: receipt.blockHash });
    assert.deepEqual(
      [stored.app, stored.version, stored.scanId, stored.artifactHash, stored.sbomHash, stored.scanHash],
      record(),
    );
    assert.equal(stored.signer.toLowerCase(), oracle.account.address);
    assert.equal(stored.timestamp, block.timestamp);
    assert.equal(await registry.read.recordCount(), 1n);
    assert.equal(await registry.read.isScanIdUsed(["scan-1"]), true);
  });

  it("emits the new record ID and unambiguous app/version key", async () => {
    const { registry, oracle, receipt } = await networkHelpers.loadFixture(recordFixture);
    const key = keccak256(encodeAbiParameters(
      [{ type: "string" }, { type: "string" }], ["demo-lib", "1.0.0"],
    ));
    await viem.assertions.emitWithArgs(receipt.transactionHash, registry, "RecordAdded", [
      1n, key, oracle.account.address, "demo-lib", "1.0.0", "scan-1",
    ]);
  });

  it("rejects writes from non-oracles", async () => {
    const { registry, stranger, asStranger } = await networkHelpers.loadFixture(deployFixture);
    await viem.assertions.revertWithCustomErrorWithArgs(
      registry.write.addRecord(record(), asStranger), registry, "UnauthorizedOracle", [stranger.account.address],
    );
    assert.equal(await registry.read.recordCount(), 0n);
    assert.equal(await registry.read.isScanIdUsed(["scan-1"]), false);
  });

  for (const operation of ["authorizeOracle", "revokeOracle"] as const) {
    it(`rejects ${operation} from non-admins`, async () => {
      const { registry, oracle, stranger, asStranger } = await networkHelpers.loadFixture(deployFixture);
      await viem.assertions.revertWithCustomErrorWithArgs(
        registry.write[operation]([oracle.account.address], asStranger), registry,
        "UnauthorizedAdmin", [stranger.account.address],
      );
    });

    it(`rejects the zero address in ${operation}`, async () => {
      const { registry } = await networkHelpers.loadFixture(deployFixture);
      await viem.assertions.revertWithCustomError(
        registry.write[operation]([zeroAddress]), registry, "ZeroAddress",
      );
    });
  }

  it("emits authorization and revocation events and rejects redundant changes", async () => {
    const { registry, oracle, stranger } = await networkHelpers.loadFixture(deployFixture);
    await viem.assertions.emitWithArgs(
      registry.write.authorizeOracle([stranger.account.address]), registry,
      "OracleAuthorizationChanged", [stranger.account.address, true],
    );
    await viem.assertions.revertWithCustomErrorWithArgs(
      registry.write.authorizeOracle([oracle.account.address]), registry,
      "OracleAlreadyAuthorized", [oracle.account.address],
    );
    await viem.assertions.emitWithArgs(
      registry.write.revokeOracle([oracle.account.address]), registry,
      "OracleAuthorizationChanged", [oracle.account.address, false],
    );
    assert.equal(await registry.read.isAuthorizedOracle([oracle.account.address]), false);
    await viem.assertions.revertWithCustomErrorWithArgs(
      registry.write.revokeOracle([oracle.account.address]), registry,
      "OracleNotAuthorized", [oracle.account.address],
    );
  });

  for (const [index, field, maximum] of [[0, "app", 214], [1, "version", 128], [2, "scanId", 128]] as const) {
    it(`rejects an empty ${field} without consuming a record or scan ID`, async () => {
      const { registry, asOracle } = await networkHelpers.loadFixture(deployFixture);
      const input: [string, string, string, Hex, Hex, Hex] = [...record()];
      input[index] = "";
      await viem.assertions.revertWithCustomErrorWithArgs(
        registry.write.addRecord(input, asOracle), registry, "EmptyField", [field],
      );
      assert.equal(await registry.read.recordCount(), 0n);
      assert.equal(await registry.read.isScanIdUsed(["scan-1"]), false);
    });

    it(`rejects ${field} above its byte limit including UTF-8 input`, async () => {
      const { registry, asOracle } = await networkHelpers.loadFixture(deployFixture);
      for (const value of ["a".repeat(maximum + 1), "é".repeat(maximum / 2 + 1)]) {
        const input: [string, string, string, Hex, Hex, Hex] = [...record()];
        input[index] = value;
        await viem.assertions.revertWithCustomErrorWithArgs(
          registry.write.addRecord(input, asOracle), registry, "StringTooLong",
          [field, BigInt(Buffer.byteLength(value)), BigInt(maximum)],
        );
      }
    });
  }

  it("accepts strings exactly at the byte limits", async () => {
    const { registry, asOracle } = await networkHelpers.loadFixture(deployFixture);
    const input = ["a".repeat(214), "v".repeat(128), "s".repeat(128), artifactHash, sbomHash, scanHash] as const;
    await registry.write.addRecord(input, asOracle);
    const stored = await registry.read.getRecord([1n]);
    assert.deepEqual([stored.app, stored.version, stored.scanId], input.slice(0, 3));
  });

  for (const [index, field] of [[3, "artifactHash"], [4, "sbomHash"], [5, "scanHash"]] as const) {
    it(`rejects a zero ${field}`, async () => {
      const { registry, asOracle } = await networkHelpers.loadFixture(deployFixture);
      const input: [string, string, string, Hex, Hex, Hex] = [...record()];
      input[index] = zeroHash;
      await viem.assertions.revertWithCustomErrorWithArgs(
        registry.write.addRecord(input, asOracle), registry, "EmptyHash", [field],
      );
    });
  }

  it("rejects scan ID replay across applications, versions, and oracles", async () => {
    const { registry, stranger } = await networkHelpers.loadFixture(recordFixture);
    await registry.write.authorizeOracle([stranger.account.address]);
    await viem.assertions.revertWithCustomErrorWithArgs(
      registry.write.addRecord(record("scan-1", "other-app", "2.0.0"), { account: stranger.account }),
      registry, "ScanIdAlreadyUsed", ["scan-1"],
    );
    assert.equal(await registry.read.recordCount(), 1n);
  });

  it("preserves records and used scan IDs after revocation and reauthorization", async () => {
    const { registry, oracle, asOracle } = await networkHelpers.loadFixture(recordFixture);
    const original = await registry.read.getRecord([1n]);
    await registry.write.revokeOracle([oracle.account.address]);
    await viem.assertions.revertWithCustomErrorWithArgs(
      registry.write.addRecord(record("scan-2"), asOracle), registry, "UnauthorizedOracle", [oracle.account.address],
    );
    assert.deepEqual(await registry.read.getRecord([1n]), original);
    assert.equal(await registry.read.isScanIdUsed(["scan-1"]), true);
    await registry.write.authorizeOracle([oracle.account.address]);
    await viem.assertions.revertWithCustomErrorWithArgs(
      registry.write.addRecord(record(), asOracle), registry, "ScanIdAlreadyUsed", ["scan-1"],
    );
  });

  it("appends rescans and changes latest without modifying older records", async () => {
    const { registry, asOracle } = await networkHelpers.loadFixture(recordFixture);
    const original = await registry.read.getRecord([1n]);
    await registry.write.addRecord(record("scan-2"), asOracle);
    assert.equal((await registry.read.getLatestRecord(["demo-lib", "1.0.0"])).scanId, "scan-2");
    assert.deepEqual(await registry.read.getRecord([1n]), original);
    assert.deepEqual(await registry.read.getRecordIds(["demo-lib", "1.0.0", 0n, 100n]), [1n, 2n]);
  });

  it("keeps ambiguous concatenations and different versions in separate histories", async () => {
    const { registry, asOracle } = await networkHelpers.loadFixture(recordFixture);
    await registry.write.addRecord(record("scan-2", "ab", "c"), asOracle);
    await registry.write.addRecord(record("scan-3", "a", "bc"), asOracle);
    await registry.write.addRecord(record("scan-4", "demo-lib", "2.0.0"), asOracle);
    assert.equal((await registry.read.getLatestRecord(["ab", "c"])).scanId, "scan-2");
    assert.equal((await registry.read.getLatestRecord(["a", "bc"])).scanId, "scan-3");
    assert.equal((await registry.read.getLatestRecord(["demo-lib", "1.0.0"])).scanId, "scan-1");
    assert.deepEqual(await registry.read.getRecordIds(["demo-lib", "2.0.0", 0n, 100n]), [4n]);
  });

  it("paginates oldest first and handles empty history and extreme offsets", async () => {
    const { registry, asOracle } = await networkHelpers.loadFixture(recordFixture);
    await registry.write.addRecord(record("scan-2"), asOracle);
    await registry.write.addRecord(record("scan-3"), asOracle);
    assert.deepEqual(await registry.read.getRecordIds(["demo-lib", "1.0.0", 0n, 2n]), [1n, 2n]);
    assert.deepEqual(await registry.read.getRecordIds(["demo-lib", "1.0.0", 2n, 2n]), [3n]);
    for (const offset of [3n, 4n, maxUint256]) {
      assert.deepEqual(await registry.read.getRecordIds(["demo-lib", "1.0.0", offset, 100n]), []);
    }
    assert.deepEqual(await registry.read.getRecordIds(["missing", "1.0.0", 0n, 100n]), []);
  });

  it("rejects zero and oversized page limits", async () => {
    const { registry } = await networkHelpers.loadFixture(recordFixture);
    for (const limit of [0n, 101n, maxUint256]) {
      await viem.assertions.revertWithCustomErrorWithArgs(
        registry.read.getRecordIds(["demo-lib", "1.0.0", 0n, limit]), registry, "InvalidPageLimit", [limit],
      );
    }
  });

  it("rejects missing records and missing latest history", async () => {
    const { registry } = await networkHelpers.loadFixture(recordFixture);
    for (const id of [0n, 2n, maxUint256]) {
      await viem.assertions.revertWithCustomErrorWithArgs(
        registry.read.getRecord([id]), registry, "RecordNotFound", [id],
      );
      await viem.assertions.revertWithCustomErrorWithArgs(
        registry.write.requestVerification([id]), registry, "RecordNotFound", [id],
      );
    }
    await viem.assertions.revertWithCustomErrorWithArgs(
      registry.read.getLatestRecord(["missing", "1.0.0"]), registry,
      "NoRecordsForAppVersion", ["missing", "1.0.0"],
    );
    assert.equal(await registry.read.verificationRequestCount(), 0n);
  });

  it("creates a pending request with the requester and block time and emits an event", async () => {
    const { registry, stranger, requestReceipt } = await networkHelpers.loadFixture(requestFixture);
    const request = await registry.read.getVerification([1n]);
    assert.equal(request.recordId, 1n);
    assert.equal(request.requester.toLowerCase(), stranger.account.address);
    assert.equal(request.requestedAt, (await publicClient.getBlock({ blockHash: requestReceipt.blockHash })).timestamp);
    assert.equal(request.fulfilled, false);
    assert.deepEqual(request.result, { status: 0, stale: false });
    assert.equal(request.fulfilledBy, zeroAddress);
    assert.equal(request.fulfilledAt, 0n);
    await viem.assertions.emitWithArgs(
      requestReceipt.transactionHash, registry, "VerificationRequested", [1n, 1n, stranger.account.address],
    );
  });

  it("fulfills once with oracle identity and block time without changing the record", async () => {
    const { registry, oracle, asOracle } = await networkHelpers.loadFixture(requestFixture);
    const original = await registry.read.getRecord([1n]);
    const result = { status: 1, stale: true };
    const receipt = await publicClient.waitForTransactionReceipt({
      hash: await registry.write.fulfillVerification([1n, result], asOracle),
    });
    const response = await registry.read.getVerification([1n]);
    assert.equal(response.fulfilled, true);
    assert.equal(response.fulfilledBy.toLowerCase(), oracle.account.address);
    assert.equal(response.fulfilledAt, (await publicClient.getBlock({ blockHash: receipt.blockHash })).timestamp);
    assert.deepEqual(response.result, result);
    assert.deepEqual(await registry.read.getRecord([1n]), original);
    await viem.assertions.emitWithArgs(receipt.transactionHash, registry, "VerificationFulfilled", [
      1n, 1n, oracle.account.address, 1, true,
    ]);
    await viem.assertions.revertWithCustomErrorWithArgs(
      registry.write.fulfillVerification([1n, result], asOracle), registry, "VerificationAlreadyFulfilled", [1n],
    );
  });

  it("rejects fulfillment by outsiders and revoked oracles", async () => {
    const { registry, oracle, stranger, asOracle, asStranger } = await networkHelpers.loadFixture(requestFixture);
    await viem.assertions.revertWithCustomErrorWithArgs(
      registry.write.fulfillVerification([1n, { status: 1, stale: false }], asStranger), registry,
      "UnauthorizedOracle", [stranger.account.address],
    );
    await registry.write.revokeOracle([oracle.account.address]);
    await viem.assertions.revertWithCustomErrorWithArgs(
      registry.write.fulfillVerification([1n, { status: 1, stale: false }], asOracle), registry,
      "UnauthorizedOracle", [oracle.account.address],
    );
    assert.equal((await registry.read.getVerification([1n])).fulfilled, false);
  });

  it("rejects missing verification requests", async () => {
    const { registry, asOracle } = await networkHelpers.loadFixture(requestFixture);
    for (const id of [0n, 2n, maxUint256]) {
      await viem.assertions.revertWithCustomErrorWithArgs(
        registry.read.getVerification([id]), registry, "VerificationRequestNotFound", [id],
      );
      await viem.assertions.revertWithCustomErrorWithArgs(
        registry.write.fulfillVerification([id, { status: 1, stale: false }], asOracle), registry,
        "VerificationRequestNotFound", [id],
      );
    }
  });

  it("rejects Pending results and stale flags on non-valid results", async () => {
    const { registry, asOracle } = await networkHelpers.loadFixture(requestFixture);
    for (const result of [{ status: 0, stale: false }, ...[2, 3, 4, 5, 6, 7].map((status) => ({ status, stale: true }))]) {
      await viem.assertions.revertWithCustomError(
        registry.write.fulfillVerification([1n, result], asOracle), registry, "InvalidVerificationResult",
      );
    }
    assert.equal((await registry.read.getVerification([1n])).fulfilled, false);
  });

  it("supports all final integrity results on independent requests", async () => {
    const { registry, asOracle, asStranger } = await networkHelpers.loadFixture(recordFixture);
    for (const status of [1, 2, 3, 4, 5, 6, 7]) {
      await registry.write.requestVerification([1n], asStranger);
      const id = await registry.read.verificationRequestCount();
      await registry.write.fulfillVerification([id, { status, stale: false }], asOracle);
      assert.deepEqual((await registry.read.getVerification([id])).result, { status, stale: false });
    }
  });
});
