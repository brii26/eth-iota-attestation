export type Selection = { network: string; app: string; version: string };

export type RecordItem = {
  recordId: string;
  scanId: string;
  status: string;
  txRef: string;
  submittedAt: string;
  includedAt?: string;
  artifactHash: string;
  sbomHash: string;
  scanHash: string;
};

export type History = { total: number; records: RecordItem[] };

export type IntegrityStatus =
  | "VALID"
  | "ARTIFACT_MISMATCH"
  | "SBOM_MISMATCH"
  | "SCAN_MISMATCH"
  | "UNTRUSTED_SIGNER"
  | "RECORD_NOT_FOUND"
  | "FILE_UNAVAILABLE";

export type Verification = Selection & {
  status: IntegrityStatus;
  recordId?: string;
  scanId?: string;
  signer?: string;
  recordedAt?: string;
  grypeDbBuiltAt?: string;
  stale: boolean;
  checks: { artifact: string; sbom: string; scan: string; signerTrusted: boolean };
  recordHashes?: { artifact: string; sbom: string; scan: string };
};

async function json<T>(path: string, signal: AbortSignal, body?: unknown): Promise<T> {
  const response = await fetch(path, {
    signal,
    ...(body === undefined ? {} : {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    }),
  });
  if (!response.ok) {
    const problem = await response.json().catch(() => undefined);
    throw new Error(problem?.error?.message || `Oracle API returned HTTP ${response.status}`);
  }
  return response.json();
}

export function networks(signal: AbortSignal) {
  return json<{ networks: string[] }>("/api/v1/networks", signal);
}

export function history(selection: Selection, offset: number, limit: number, signal: AbortSignal) {
  const query = new URLSearchParams({ ...selection, offset: String(offset), limit: String(limit) });
  return json<History>(`/api/v1/records?${query}`, signal);
}

export function verify(selection: Selection, recordId: string | undefined, signal: AbortSignal) {
  return json<Verification>("/api/v1/verify", signal, { ...selection, recordId });
}

export const statusText: Record<IntegrityStatus, { label: string; explanation: string }> = {
  VALID: { label: "Valid", explanation: "All three files match the record, and the recording oracle is trusted." },
  ARTIFACT_MISMATCH: { label: "Package changed", explanation: "The downloaded package does not match the recorded fingerprint." },
  SBOM_MISMATCH: { label: "SBOM changed", explanation: "The software bill of materials does not match the recorded fingerprint." },
  SCAN_MISMATCH: { label: "Scan changed", explanation: "The vulnerability scan does not match the recorded fingerprint." },
  UNTRUSTED_SIGNER: { label: "Untrusted oracle", explanation: "The oracle that wrote this record is no longer trusted by the selected network." },
  RECORD_NOT_FOUND: { label: "Record not found", explanation: "The selected release or record has no attestation on this network." },
  FILE_UNAVAILABLE: { label: "Evidence unavailable", explanation: "The ledger record or an off-chain file could not be retrieved." },
};
