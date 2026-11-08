import { readFile } from "node:fs/promises";
import { createPublicClient, createWalletClient, defineChain, http } from "viem";
import { privateKeyToAccount } from "viem/accounts";
import type { Hex } from "viem";
import type { AttestationRegistry$Type } from "../artifacts/contracts/AttestationRegistry.sol/artifacts.d.ts";

export const artifact = JSON.parse(await readFile(
  new URL("../artifacts/contracts/AttestationRegistry.sol/AttestationRegistry.json", import.meta.url), "utf8",
)) as Pick<AttestationRegistry$Type, "abi" | "bytecode">;

export const deploymentFile = process.env.DEPLOYMENT_FILE ?? "deployments/ethereum-private.json";

export async function privateNetwork() {
  const url = process.env.ETHEREUM_RPC_URL ?? "http://127.0.0.1:8545";
  const chain = defineChain({
    id: 1337,
    name: "Private Geth",
    nativeCurrency: { name: "Ether", symbol: "ETH", decimals: 18 },
    rpcUrls: { default: { http: [url] } },
  });
  const transport = http(url, { timeout: 15_000, retryCount: 0 });
  const publicClient = createPublicClient({ chain, transport });
  if (await publicClient.getChainId() !== chain.id) {
    throw new Error("Expected the private developer chain with chain ID 1337");
  }
  return { chain, transport, publicClient };
}

export function privateAccount(name: string) {
  const value = process.env[name];
  if (!value || !/^0x[0-9a-fA-F]{64}$/.test(value)) {
    throw new Error(`${name} must contain a 32-byte hex private key`);
  }
  try {
    return privateKeyToAccount(value as Hex);
  } catch {
    throw new Error(`${name} is not a valid private key`);
  }
}

export const wallet = createWalletClient;
