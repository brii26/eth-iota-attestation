import { createPublicClient, defineChain, http, isAddress, parseAbi } from "viem";
import { sepolia } from "viem/chains";
import type { Selection, Verification } from "./api";

let deployedPrivateRegistry: string | undefined;

export async function loadRegistryDeployment() {
  if (import.meta.env.VITE_PRIVATE_REGISTRY) return;
  try {
    const response = await fetch("/deployment.json", { signal: AbortSignal.timeout(3_000) });
    if (!response.ok || !response.headers.get("Content-Type")?.includes("application/json")) return;
    const deployment = await response.json();
    if (deployment.chainId === 1337 && typeof deployment.contractAddress === "string" &&
        isAddress(deployment.contractAddress)) {
      deployedPrivateRegistry = deployment.contractAddress;
    }
  } catch {
    return;
  }
}

const privateRPC = import.meta.env.VITE_PRIVATE_RPC_URL || "http://127.0.0.1:8545";
const privateChain = defineChain({
  id: 1337,
  name: "Private Ethereum",
  nativeCurrency: { name: "Ether", symbol: "ETH", decimals: 18 },
  rpcUrls: { default: { http: [privateRPC] } },
});

const recordTuple = "(string app, string version, string scanId, bytes32 artifactHash, bytes32 sbomHash, bytes32 scanHash, address signer, uint256 timestamp)";
export const registryABI = parseAbi([
  `function getLatestRecord(string app, string version) view returns (${recordTuple})`,
  `function getRecord(uint256 recordId) view returns (${recordTuple})`,
]);

export function ethereumNetwork(network: string) {
  let address: string | undefined;
  let rpc: string | undefined;
  let chain;
  if (network === "ethereum-private") {
    address = import.meta.env.VITE_PRIVATE_REGISTRY || deployedPrivateRegistry;
    rpc = privateRPC;
    chain = privateChain;
  } else if (network === "ethereum-sepolia") {
    address = import.meta.env.VITE_SEPOLIA_REGISTRY;
    rpc = import.meta.env.VITE_SEPOLIA_RPC_URL;
    chain = sepolia;
  }
  if (!address || !isAddress(address) || !chain) return undefined;
  return { address, chain, client: createPublicClient({ chain, transport: http(rpc, { timeout: 8_000 }) }) };
}

export async function compareWithRegistry(selection: Selection, selectedId: string | undefined, result: Verification) {
  const target = ethereumNetwork(selection.network);
  if (!target) throw new Error("Set the registry address for this network in the dashboard configuration.");
  const { client, address, chain } = target;
  if (await client.getChainId() !== chain.id) throw new Error("The configured RPC is on a different chain.");
  const contract = { address, abi: registryABI };
  const record = selectedId
    ? await client.readContract({ ...contract, functionName: "getRecord", args: [BigInt(selectedId)] })
    : await client.readContract({ ...contract, functionName: "getLatestRecord", args: [selection.app, selection.version] });

  return record.app === result.app &&
    record.version === result.version &&
    record.scanId === result.scanId &&
    record.signer.toLowerCase() === result.signer?.toLowerCase() &&
    Number(record.timestamp) * 1000 === Date.parse(result.recordedAt ?? "") &&
    record.artifactHash.toLowerCase() === result.recordHashes?.artifact &&
    record.sbomHash.toLowerCase() === result.recordHashes?.sbom &&
    record.scanHash.toLowerCase() === result.recordHashes?.scan;
}
