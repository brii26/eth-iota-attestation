import { createServer } from "node:http";
import { timingSafeEqual } from "node:crypto";
import type { IncomingMessage, ServerResponse } from "node:http";
import { IotaClient, getFullnodeUrl, JsonRpcError } from "@iota/iota-sdk/client";
import type { IotaTransactionBlockResponse } from "@iota/iota-sdk/client";
import { Ed25519Keypair } from "@iota/iota-sdk/keypairs/ed25519";
import { TransactionDataBuilder } from "@iota/iota-sdk/transactions";
import { iota_client_helpers } from "@iota/iota-interaction-ts/node/index.js";
import { CreateNotarizationLocked, NotarizationBuilderLocked, NotarizationClientReadOnly, TimeLock } from "@iota/notarization/node/index.js";

const marker = "artifact-attestation-v1";
const objectID = /^0x[0-9a-f]{64}$/;
const digestFormat = /^[1-9A-HJ-NP-Za-km-z]{32,44}$/;
const options = { showEffects: true, showInput: true, showObjectChanges: true, showRawInput: true };

type RecordPayload = {
  schemaVersion: 1;
  app: string;
  version: string;
  scanId: string;
  artifactHash: string;
  sbomHash: string;
  scanHash: string;
};

class RequestError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

function payload(value: unknown): RecordPayload {
  if (typeof value !== "object" || value === null) throw new RequestError(400, "invalid attestation");
  const candidate = value as RecordPayload;
  if (
    candidate.schemaVersion !== 1 ||
    ![candidate.app, candidate.version, candidate.scanId].every(
      (field) => typeof field === "string" && field.length > 0 && Buffer.byteLength(field) <= 256,
    ) ||
    ![candidate.artifactHash, candidate.sbomHash, candidate.scanHash].every(
      (hash) => typeof hash === "string" && objectID.test(hash),
    )
  ) throw new RequestError(400, "invalid attestation");
  const { schemaVersion, app, version, scanId, artifactHash, sbomHash, scanHash } = candidate;
  return { schemaVersion, app, version, scanId, artifactHash, sbomHash, scanHash };
}

async function body(req: IncomingMessage): Promise<unknown> {
  const chunks: Buffer[] = [];
  let length = 0;
  for await (const chunk of req) {
    const bytes = Buffer.from(chunk);
    length += bytes.length;
    if (length > 128 * 1024) throw new RequestError(413, "request too large");
    chunks.push(bytes);
  }
  try {
    return JSON.parse(Buffer.concat(chunks).toString("utf8"));
  } catch {
    throw new RequestError(400, "invalid JSON");
  }
}

const network = process.env.IOTA_NETWORK ?? "testnet";
if (network !== "testnet" && network !== "localnet") throw new Error("IOTA_NETWORK must be testnet or localnet");
const port = Number(process.env.IOTA_SERVICE_PORT ?? "8091");
const lockSeconds = Number(process.env.IOTA_DELETE_LOCK_SECONDS ?? "31536000");
if (!Number.isSafeInteger(lockSeconds) || lockSeconds < 60 || lockSeconds > 315360000) {
  throw new Error("IOTA_DELETE_LOCK_SECONDS must be between 60 and 315360000");
}
const token = Buffer.from(process.env.IOTA_SERVICE_TOKEN ?? "");
if (token.length < 32) throw new Error("IOTA_SERVICE_TOKEN must be at least 32 bytes");
const privateKey = process.env.IOTA_PRIVATE_KEY;
if (!privateKey) throw new Error("IOTA_PRIVATE_KEY is required");
const signer = Ed25519Keypair.fromSecretKey(
  /^(0x)?[0-9a-fA-F]{64}$/.test(privateKey) ? Buffer.from(privateKey.replace(/^0x/, ""), "hex") : privateKey,
);
const address = signer.toIotaAddress();
const rpcURL = process.env.IOTA_RPC_URL ?? getFullnodeUrl(network);
const rpc = new IotaClient({ url: rpcURL });
const officialURL = getFullnodeUrl("testnet");
if (network === "testnet" && rpcURL !== officialURL) {
  const identifiers = await Promise.all([rpc.getChainIdentifier(), new IotaClient({ url: officialURL }).getChainIdentifier()]);
  if (identifiers[0] !== identifiers[1]) throw new Error("IOTA RPC does not match testnet");
}
const packageID = process.env.IOTA_NOTARIZATION_PKG_ID;
type NotarizationRPC = Parameters<typeof NotarizationClientReadOnly.create>[0];
const notarizationRPC = rpc as unknown as NotarizationRPC;
const readOnly = packageID
  ? await NotarizationClientReadOnly.createWithPkgId(notarizationRPC, packageID)
  : await NotarizationClientReadOnly.create(notarizationRPC);

async function prepare(record: RecordPayload) {
  const sequence = await rpc.getLatestCheckpointSequenceNumber();
  const checkpoint = await rpc.getCheckpoint({ id: sequence });
  const deleteUnlockAt = Math.floor(Number(checkpoint.timestampMs) / 1000) + lockSeconds;
  const builder = NotarizationBuilderLocked.locked()
    .withStringState(JSON.stringify(record), marker)
    .withImmutableDescription("Software artifact attestation")
    .withDeleteLock(TimeLock.withUnlockAt(deleteUnlockAt));
  const transaction = new CreateNotarizationLocked(builder);
  type BuildClient = Parameters<typeof transaction.buildProgrammableTransaction>[0];
  const unsigned = await transaction.buildProgrammableTransaction(readOnly as unknown as BuildClient);
  const bytes = await iota_client_helpers.addGasDataToTransaction(
    rpc as unknown as Parameters<typeof iota_client_helpers.addGasDataToTransaction>[0], address, unsigned,
  );
  const signed = await signer.signTransaction(bytes);
  return { digest: TransactionDataBuilder.getDigestFromBytes(bytes), bytes: signed.bytes, signature: signed.signature };
}

async function fetchTransaction(digest: string) {
  try {
    return await rpc.getTransactionBlock({ digest, options });
  } catch (error) {
    if (error instanceof JsonRpcError && error.code === -32003 && /could not find.*transaction/i.test(error.message)) {
      return undefined;
    }
    throw error;
  }
}

async function receipt(tx: IotaTransactionBlockResponse, expectedObject?: string) {
  if (!tx.effects) throw new RequestError(503, "transaction effects are not available");
  if (tx.effects.status.status === "failure") throw new RequestError(422, "IOTA transaction failed");
  const object = tx.objectChanges?.find(
    (change) => change.type === "created" && change.objectType.startsWith(readOnly.packageId() + "::") &&
      (!expectedObject || change.objectId === expectedObject),
  );
  if (object?.type !== "created" || !tx.transaction?.data.sender) {
    throw new RequestError(422, "transaction did not create an attestation object");
  }
  const timestamp = tx.timestampMs ? new Date(Number(tx.timestampMs)).toISOString() : undefined;
  const gas = tx.effects.gasUsed;
  return {
    recordId: object.objectId, txRef: tx.digest, signer: tx.transaction.data.sender,
    includedAt: timestamp, finalizedAt: tx.checkpoint ? timestamp : undefined,
    checkpoint: tx.checkpoint, gasComputation: gas.computationCost, gasStorage: gas.storageCost,
    gasRebate: gas.storageRebate, txSize: tx.rawTransaction ? Buffer.from(tx.rawTransaction, "base64").length : 0,
  };
}

async function submit(value: unknown) {
  const request = value as { digest?: string; bytes?: string; signature?: string } | null;
  if (!request?.digest || !digestFormat.test(request.digest) ||
    typeof request.bytes !== "string" || typeof request.signature !== "string") {
    throw new RequestError(400, "invalid prepared transaction");
  }
  const bytes = Buffer.from(request.bytes, "base64");
  if (TransactionDataBuilder.getDigestFromBytes(bytes) !== request.digest ||
      TransactionDataBuilder.fromBytes(bytes).sender !== address) {
    throw new RequestError(400, "prepared transaction does not match signer or digest");
  }
  const known = await fetchTransaction(request.digest);
  const tx = known ?? await rpc.executeTransactionBlock({ transactionBlock: bytes, signature: request.signature, options });
  const result = await receipt(tx);
  if (result.includedAt) return { ...result, txSize: bytes.length };
  const confirmed = await rpc.waitForTransaction({ digest: tx.digest, options, timeout: 12_000 });
  return { ...await receipt(confirmed), txSize: bytes.length };
}

async function read(id: string) {
  if (!objectID.test(id)) throw new RequestError(400, "invalid IOTA object ID");
  const object = await rpc.getObject({ id, options: { showPreviousTransaction: true } });
  if (object.error?.code === "notExists" || object.error?.code === "deleted") {
    throw new RequestError(404, "notarization object is unavailable");
  }
  const digest = object.data?.previousTransaction;
  if (!digest) throw new RequestError(503, "notarization transaction is unavailable");
  const notarization = await readOnly.getNotarizationById(id);
  if (notarization.method !== "Locked" || notarization.state.metadata !== marker) {
    throw new RequestError(404, "object is not a locked attestation");
  }
  const tx = await fetchTransaction(digest);
  if (!tx) throw new RequestError(503, "notarization creation transaction is unavailable");
  const record = payload(JSON.parse(notarization.state.data.toString()));
  const lock = notarization.immutableMetadata.locking?.deleteLock;
  const deleteUnlockAt = lock?.type === "UnlockAt" && typeof lock.args === "number"
    ? new Date(lock.args * 1000).toISOString() : undefined;
  return {
    ...record, ...await receipt(tx, id), deleteUnlockAt,
    timestamp: new Date(Number(notarization.immutableMetadata.createdAt)).toISOString(),
  };
}

async function list(search: URLSearchParams) {
  const owner = search.get("owner") ?? address;
  if (!objectID.test(owner)) throw new RequestError(400, "invalid IOTA owner");
  const cursor = search.get("cursor") ?? undefined;
  const page = await rpc.getOwnedObjects({
    owner, cursor, limit: 50,
    options: { showType: true },
  });
  return {
    ids: page.data.flatMap(({ data }) => data?.type?.startsWith(readOnly.packageId() + "::") ? [data.objectId] : []),
    cursor: page.hasNextPage ? page.nextCursor : null,
  };
}

function reply(res: ServerResponse, status: number, data: unknown) {
  res.writeHead(status, { "Content-Type": "application/json; charset=utf-8" });
  res.end(JSON.stringify(data));
}

createServer(async (req, res) => {
  try {
    const authorization = Buffer.from((req.headers.authorization ?? "").replace(/^Bearer /, ""));
    if (authorization.length !== token.length || !timingSafeEqual(authorization, token)) {
      throw new RequestError(401, "unauthorized");
    }
    const url = new URL(req.url ?? "/", "http://127.0.0.1");
    if (req.method === "POST" && url.pathname === "/prepare") {
      reply(res, 200, await prepare(payload(await body(req))));
    } else if (req.method === "POST" && url.pathname === "/submit") {
      reply(res, 200, await submit(await body(req)));
    } else if (req.method === "GET" && url.pathname.startsWith("/record/")) {
      reply(res, 200, await read(url.pathname.slice("/record/".length)));
    } else if (req.method === "GET" && url.pathname === "/records") {
      reply(res, 200, await list(url.searchParams));
    } else if (req.method === "GET" && url.pathname === "/network") {
      reply(res, 200, { name: `iota-${network}`, signer: address });
    } else {
      throw new RequestError(404, "route not found");
    }
  } catch (error) {
    if (error instanceof RequestError) reply(res, error.status, { error: error.message });
    else {
      console.error(error instanceof Error ? error.message.replaceAll(privateKey, "[redacted]") : "IOTA RPC failed");
      reply(res, 502, { error: "IOTA RPC failed" });
    }
  }
}).listen(port, "127.0.0.1", () => console.log(`IOTA ${network} service listening on 127.0.0.1:${port}`));
