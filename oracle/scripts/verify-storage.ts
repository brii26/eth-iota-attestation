import {
  S3Client,
  CreateBucketCommand,
  HeadBucketCommand,
  PutObjectCommand,
  GetObjectCommand,
  ListObjectsV2Command,
} from "@aws-sdk/client-s3";
import type { Readable } from "stream";
import * as crypto from "crypto";

// --- Configuration ---
const S3_ENDPOINT = process.env.RUSTFS_ENDPOINT || "http://127.0.0.1:9000";
const S3_REGION = process.env.AWS_REGION || "us-east-1";
const S3_CREDENTIALS = {
  accessKeyId: process.env.RUSTFS_ACCESS_KEY || "rustfsadmin",
  secretAccessKey: process.env.RUSTFS_SECRET_KEY || "rustfsdevpassword123",
};

const VERDACCIO_URL = process.env.VERDACCIO_URL || "http://127.0.0.1:4873";
const TEST_PACKAGE_NAME = `@test/artifact-${Date.now()}`;
const TEST_VERSION = "1.0.0";

// --- Types ---
interface ObjectVerificationResult {
  verified: boolean;
  sha256: string;
  bytes: number;
  listed: boolean;
}

interface S3TestResult {
  buckets: Record<string, string>;
  objects: Record<string, ObjectVerificationResult>;
}

interface RejectionResult {
  status: number;
  asserted: boolean;
}

interface VerdaccioTestResult {
  firstPublish: { status: number; success: boolean };
  duplicatePublishRejection: RejectionResult;
  unpublishRejection: RejectionResult;
}

interface SuiteOutput {
  timestamp: string;
  status: "PASSED" | "FAILED";
  tests: {
    s3?: S3TestResult;
    verdaccio?: VerdaccioTestResult;
  };
  error?: string;
}

interface CouchDbAuthResponse {
  ok: boolean;
  id: string;
  rev: string;
  token: string;
}

// --- Helpers ---
function calculateSHA256(data: Buffer): string {
  return crypto.createHash("sha256").update(data).digest("hex");
}

function calculateSHA1(data: Buffer): string {
  return crypto.createHash("sha1").update(data).digest("hex");
}

async function streamToBuffer(stream: Readable): Promise<Buffer> {
  const chunks: Buffer[] = [];
  for await (const chunk of stream) {
    chunks.push(Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk));
  }
  return Buffer.concat(chunks);
}

// 1. RustFS / S3 Verification
async function verifyS3(): Promise<S3TestResult> {
  const results: S3TestResult = { buckets: {}, objects: {} };

  const s3 = new S3Client({
    endpoint: S3_ENDPOINT,
    region: S3_REGION,
    credentials: S3_CREDENTIALS,
    forcePathStyle: true, // Path-style access (e.g. http://127.0.0.1:9000/bucket)
  });

  const buckets = ["sboms", "scans"];

  // Step 1: Ensure Buckets Exist
  for (const bucket of buckets) {
    try {
      await s3.send(new HeadBucketCommand({ Bucket: bucket }));
    } catch {
      await s3.send(new CreateBucketCommand({ Bucket: bucket }));
    }
    results.buckets[bucket] = "exists_or_created";
  }

  // Step 2: Upload, Read Back, and Verify Integrity
  const testFiles = [
    {
      bucket: "sboms",
      key: "sample-sbom.json",
      content: Buffer.from(JSON.stringify({ name: "app", version: "1.0.0", dependencies: {} })),
    },
    {
      bucket: "scans",
      key: "sample-scan.json",
      content: Buffer.from(JSON.stringify({ scanId: "12345", vulnerabilities: [] })),
    },
  ];

  for (const file of testFiles) {
    const originalHash = calculateSHA256(file.content);

    // Put Object
    await s3.send(
      new PutObjectCommand({
        Bucket: file.bucket,
        Key: file.key,
        Body: file.content,
        ContentType: "application/json",
      })
    );

    // Get Object
    const getResponse = await s3.send(
      new GetObjectCommand({
        Bucket: file.bucket,
        Key: file.key,
      })
    );

    if (!getResponse.Body) {
      throw new Error(`Empty body returned for ${file.bucket}/${file.key}`);
    }

    const retrievedBuffer = await streamToBuffer(getResponse.Body as Readable);
    const retrievedHash = calculateSHA256(retrievedBuffer);

    // Assert SHA-256 & byte-for-byte match
    const hashMatch = originalHash === retrievedHash;
    const byteMatch = file.content.equals(retrievedBuffer);

    if (!hashMatch || !byteMatch) {
      throw new Error(`Data corruption detected for ${file.bucket}/${file.key}`);
    }

    // Step 3: List Objects & Confirm Presence
    const listResponse = await s3.send(new ListObjectsV2Command({ Bucket: file.bucket }));
    const existsInList = listResponse.Contents?.some((obj) => obj.Key === file.key) ?? false;

    if (!existsInList) {
      throw new Error(`Key ${file.key} not found in ListObjectsV2 response for bucket ${file.bucket}`);
    }

    results.objects[`${file.bucket}/${file.key}`] = {
      verified: true,
      sha256: retrievedHash,
      bytes: retrievedBuffer.length,
      listed: true,
    };
  }

  return results;
}

// 2. Verdaccio Registry Verification
async function verifyVerdaccio(): Promise<VerdaccioTestResult> {
  const username = `testuser_${Date.now()}`;
  const password = "Password123!";
  const email = `${username}@example.com`;

  // Step 1: Register User / Obtain Auth Token
  const userUrl = `${VERDACCIO_URL}/-/user/org.couchdb.user:${username}`;
  const regRes = await fetch(userUrl, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      name: username,
      password,
      email,
      type: "user",
      roles: [],
      date: new Date().toISOString(),
    }),
  });

  if (!regRes.ok && regRes.status !== 201 && regRes.status !== 200) {
    throw new Error(`Failed to register Verdaccio user: ${regRes.status} ${await regRes.text()}`);
  }

  const regData = (await regRes.json()) as CouchDbAuthResponse;
  const token = regData.token;
  const authHeader = { Authorization: `Bearer ${token}` };

  // Helper to build a valid npm publish payload
  function createNpmPayload(name: string, version: string) {
    const tarballContent = Buffer.from("mock package content");
    const tarballBase64 = tarballContent.toString("base64");
    const scopeTrimmedName = name.replace(/^@.*\//, "");
    const filename = `${scopeTrimmedName}-${version}.tgz`;

    return {
      _id: name,
      name,
      "dist-tags": { latest: version },
      versions: {
        [version]: {
          name,
          version,
          dist: {
            shasum: calculateSHA1(tarballContent),
            tarball: `${VERDACCIO_URL}/${name}/-/${filename}`,
          },
        },
      },
      _attachments: {
        [filename]: {
          content_type: "application/octet-stream",
          data: tarballBase64,
          length: tarballContent.length,
        },
      },
    };
  }

  const pkgUrl = `${VERDACCIO_URL}/${encodeURIComponent(TEST_PACKAGE_NAME)}`;
  const publishPayload = createNpmPayload(TEST_PACKAGE_NAME, TEST_VERSION);

  // Step 2: Publish mock package (first attempt)
  const pubRes = await fetch(pkgUrl, {
    method: "PUT",
    headers: { ...authHeader, "Content-Type": "application/json" },
    body: JSON.stringify(publishPayload),
  });

  if (!pubRes.ok) {
    throw new Error(`First publish failed unexpectedly: ${pubRes.status} ${await pubRes.text()}`);
  }

  // Step 3: Attempt duplicate publish (Assert rejection: 409 Conflict or 403 Forbidden)
  const dupRes = await fetch(pkgUrl, {
    method: "PUT",
    headers: { ...authHeader, "Content-Type": "application/json" },
    body: JSON.stringify(publishPayload),
  });

  const dupRejected = dupRes.status === 409 || dupRes.status === 403;
  if (!dupRejected) {
    throw new Error(`Duplicate publish expected 409/403, got status ${dupRes.status}`);
  }

  // Step 4: Attempt unpublish (Assert rejection: 403 Forbidden / unpublish disabled)
  const unpubUrl = `${VERDACCIO_URL}/${encodeURIComponent(TEST_PACKAGE_NAME)}/-rev/1-rev`;
  const unpubRes = await fetch(unpubUrl, {
    method: "DELETE",
    headers: { ...authHeader, "Content-Type": "application/json" },
  });

  const unpubRejected = unpubRes.status === 403 || !unpubRes.ok;
  if (!unpubRejected) {
    throw new Error(`Unpublish expected rejection (403), received status ${unpubRes.status}`);
  }

  return {
    firstPublish: { status: pubRes.status, success: true },
    duplicatePublishRejection: { status: dupRes.status, asserted: dupRejected },
    unpublishRejection: { status: unpubRes.status, asserted: unpubRejected },
  };
}

// Main Runner
async function main(): Promise<void> {
  const output: SuiteOutput = {
    timestamp: new Date().toISOString(),
    status: "PASSED",
    tests: {},
  };

  try {
    output.tests.s3 = await verifyS3();
    output.tests.verdaccio = await verifyVerdaccio();
  } catch (error) {
    output.status = "FAILED";
    output.error = error instanceof Error ? error.message : String(error);
  } finally {
    console.log(JSON.stringify(output, null, 2));
    if (output.status === "FAILED") {
      process.exit(1);
    }
  }
}

main();
