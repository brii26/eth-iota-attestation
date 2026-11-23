import { useEffect, useState } from "react";
import { isAddress, parseAbi, zeroAddress } from "viem";
import type { Address } from "viem";
import type { ethereumNetwork } from "./ethereum";
import { Button } from "./components/ui/button";

const badgeABI = parseAbi([
  "function balanceOf(address owner) view returns (uint256)",
  "function tokenOfOwnerByIndex(address owner, uint256 index) view returns (uint256)",
  "function recordOf(uint256 tokenId) view returns (uint256)",
]);

const pageSize = 6n;

export function configuredBadge(network: string): Address | undefined {
  const address = network === "ethereum-private" ? import.meta.env.VITE_PRIVATE_BADGE_CONTRACT_ADDRESS :
    network === "ethereum-sepolia" ? import.meta.env.VITE_SEPOLIA_BADGE_CONTRACT_ADDRESS : undefined;
  return address && isAddress(address) && address.toLowerCase() !== zeroAddress ? address : undefined;
}

export function Badges({ target, address, account }: {
  target: NonNullable<ReturnType<typeof ethereumNetwork>>;
  address: Address;
  account: Address;
}) {
  const [offset, setOffset] = useState(0n);
  const [results, setResults] = useState<{ count: bigint; badges: { tokenId: bigint; recordId: bigint }[] }>();
  const [error, setError] = useState(false);

  useEffect(() => {
    let active = true;
    let timer: ReturnType<typeof setTimeout> | undefined;
    async function refresh() {
      try {
        const contract = { address, abi: badgeABI };
        if (await target.client.getChainId() !== target.chain.id) throw new Error("Badge RPC uses another network");
        const count = await target.client.readContract({ ...contract, functionName: "balanceOf", args: [account] });
        const length = count > offset ? Math.min(Number(count - offset), Number(pageSize)) : 0;
        const badges = await Promise.all(Array.from({ length }, async (_, index) => {
          const tokenId = await target.client.readContract({
            ...contract, functionName: "tokenOfOwnerByIndex", args: [account, offset + BigInt(index)],
          });
          const recordId = await target.client.readContract({ ...contract, functionName: "recordOf", args: [tokenId] });
          return { tokenId, recordId };
        }));
        if (!active) return;
        setResults({ count, badges });
        setError(false);
      } catch {
        if (active) setError(true);
      }
      if (active) timer = setTimeout(refresh, 5_000);
    }
    void refresh();
    return () => {
      active = false;
      clearTimeout(timer);
    };
  }, [target, address, account, offset]);

  return <div>
    <h3>Attestation badges</h3>
    {error && <p className="notice" role="status">Could not read badges. Retrying.</p>}
    {!results && !error && <p role="status">Loading badges…</p>}
    {results?.count === 0n && <p>No badges yet.</p>}
    {!!results?.badges.length && <ul>
      {results.badges.map(({ tokenId, recordId }) => <li key={tokenId.toString()}>
        Badge #{tokenId.toString()} · Record #{recordId.toString()}
      </li>)}
    </ul>}
    {results && results.count > pageSize && <div className="row">
      <Button variant="secondary" disabled={offset === 0n} onClick={() => setOffset(offset - pageSize)}>Previous badges</Button>
      <Button variant="secondary" disabled={offset + pageSize >= results.count} onClick={() => setOffset(offset + pageSize)}>Next badges</Button>
    </div>}
  </div>;
}
