import { useEffect, useMemo, useState } from "react";
import type { FormEvent } from "react";
import { createWalletClient, custom, isAddress, parseAbi, parseEventLogs, zeroAddress } from "viem";
import type { Address, EIP1193Provider, Hash } from "viem";
import { ethereumNetwork } from "./ethereum";
import { Badges, configuredBadge } from "./Badges";
import { Button } from "./components/ui/button";

const operationsABI = parseAbi([
  "function admin() view returns (address)",
  "function authorizeOracle(address oracle)",
  "function revokeOracle(address oracle)",
  "function requestVerification(uint256 recordId) returns (uint256 requestId)",
  "event VerificationRequested(uint256 indexed requestId, uint256 indexed recordId, address indexed requester)",
  "function getVerification(uint256 requestId) view returns ((uint256 recordId, address requester, uint256 requestedAt, bool fulfilled, (uint8 status, bool stale) result, address fulfilledBy, uint256 fulfilledAt))",
  "error UnauthorizedAdmin(address caller)",
  "error OracleAlreadyAuthorized(address oracle)",
  "error OracleNotAuthorized(address oracle)",
  "error RecordNotFound(uint256 recordId)",
]);
const statuses = ["Pending", "Valid", "Package changed", "SBOM changed", "Scan changed", "Untrusted oracle", "Record not found", "Evidence unavailable"];

type MetaMaskProvider = EIP1193Provider & { isMetaMask?: boolean; providers?: MetaMaskProvider[] };
type TransactionState = { phase: "signing" | "pending" | "confirmed" | "failed"; message: string; hash?: Hash };

function getProvider() {
  const injected = (window as Window & { ethereum?: MetaMaskProvider }).ethereum;
  return injected?.providers?.find((provider) => provider.isMetaMask) || (injected?.isMetaMask ? injected : undefined);
}

function errorMessage(error: unknown) {
  if (error instanceof Error) {
    if ("shortMessage" in error && typeof error.shortMessage === "string") return error.shortMessage;
    return error.message;
  }
  return "The wallet could not complete the request.";
}

function walletErrorCode(error: unknown) {
  return typeof error === "object" && error !== null && "code" in error ? error.code : undefined;
}

export function Wallet({ network, recordId }: { network: string; recordId?: string }) {
  const target = useMemo(() => ethereumNetwork(network), [network]);
  const badge = configuredBadge(network);
  const [provider, setProvider] = useState<MetaMaskProvider>();
  const [account, setAccount] = useState<Address>();
  const [oracle, setOracle] = useState("");
  const [state, setState] = useState<TransactionState>();
  const [requestId, setRequestId] = useState<bigint>();
  const [verification, setVerification] = useState<{ status: string; stale: boolean; oracle: Address }>();
  const [pollError, setPollError] = useState("");
  const busy = state?.phase === "signing" || state?.phase === "pending";

  useEffect(() => {
    if (!provider) return;
    function accountsChanged(accounts: Address[]) { setAccount(accounts[0]); }
    function disconnected() { setAccount(undefined); }
    provider.on("accountsChanged", accountsChanged);
    provider.on("disconnect", disconnected);
    return () => {
      provider.removeListener("accountsChanged", accountsChanged);
      provider.removeListener("disconnect", disconnected);
    };
  }, [provider]);

  useEffect(() => {
    if (!target || requestId === undefined) return;
    const registry = target;
    const id = requestId;
    let active = true;
    let timer: ReturnType<typeof setTimeout> | undefined;
    async function refresh() {
      try {
        const result = await registry.client.readContract({
          address: registry.address, abi: operationsABI, functionName: "getVerification", args: [id],
        });
        if (!active) return;
        setPollError("");
        if (result.fulfilled) {
          setVerification({ status: statuses[result.result.status] || "Unknown status", stale: result.result.stale, oracle: result.fulfilledBy });
          return;
        }
      } catch {
        if (active) setPollError("Could not read the request. Retrying.");
      }
      if (active) timer = setTimeout(refresh, 4_000);
    }
    void refresh();
    return () => {
      active = false;
      clearTimeout(timer);
    };
  }, [target, requestId]);

  async function switchChain(wallet: MetaMaskProvider) {
    if (!target) throw new Error("Configure the registry for this Ethereum network first.");
    const chainId = await wallet.request({ method: "eth_chainId" });
    if (Number(chainId) === target.chain.id) return;
    const chainIdHex = `0x${target.chain.id.toString(16)}` as const;
    try {
      await wallet.request({ method: "wallet_switchEthereumChain", params: [{ chainId: chainIdHex }] });
    } catch (error) {
      if (walletErrorCode(error) !== 4902) throw error;
      await wallet.request({
        method: "wallet_addEthereumChain",
        params: [{
          chainId: chainIdHex, chainName: target.chain.name,
          nativeCurrency: target.chain.nativeCurrency,
          rpcUrls: target.chain.rpcUrls.default.http.slice(),
        }],
      });
      await wallet.request({ method: "wallet_switchEthereumChain", params: [{ chainId: chainIdHex }] });
    }
    const updatedChain = await wallet.request({ method: "eth_chainId" });
    if (Number(updatedChain) !== target.chain.id) throw new Error("MetaMask is still connected to a different network.");
  }

  async function connect() {
    setState(undefined);
    const wallet = getProvider();
    if (!wallet) {
      setState({ phase: "failed", message: "MetaMask was not detected in this browser." });
      return;
    }
    try {
      const [address] = await wallet.request({ method: "eth_requestAccounts" });
      if (!address) throw new Error("MetaMask did not return an account.");
      await switchChain(wallet);
      setProvider(wallet);
      setAccount(address);
    } catch (error) {
      setState({ phase: "failed", message: errorMessage(error) });
    }
  }

  async function send(operation: "authorizeOracle" | "revokeOracle" | "requestVerification") {
    if (!provider || !account || !target) return;
    setState({ phase: "signing", message: "Confirm the transaction in MetaMask." });
    let hash: Hash | undefined;
    try {
      await switchChain(provider);
      if (await target.client.getChainId() !== target.chain.id) throw new Error("The configured RPC is on a different chain.");
      const wallet = createWalletClient({ account, chain: target.chain, transport: custom(provider) });
      const contract = { address: target.address, abi: operationsABI, account } as const;
      if (operation === "requestVerification") {
        if (!recordId || !/^\d+$/.test(recordId)) throw new Error("Select or verify an Ethereum record first.");
        const { request } = await target.client.simulateContract({ ...contract, functionName: operation, args: [BigInt(recordId)] });
        hash = await wallet.writeContract(request);
      } else {
        if (!isAddress(oracle) || oracle.toLowerCase() === zeroAddress) throw new Error("Enter a valid oracle address.");
        const admin = await target.client.readContract({ ...contract, functionName: "admin" });
        if (admin.toLowerCase() !== account.toLowerCase()) throw new Error("Only the registry administrator can change oracle permissions.");
        const { request } = await target.client.simulateContract({ ...contract, functionName: operation, args: [oracle] });
        hash = await wallet.writeContract(request);
      }
      setState({ phase: "pending", message: "Waiting for the transaction receipt.", hash });
      const receipt = await target.client.waitForTransactionReceipt({ hash, timeout: 120_000 });
      if (receipt.status !== "success") throw new Error("The transaction reverted.");
      if (operation === "requestVerification") {
        const logs = parseEventLogs({
          abi: operationsABI, eventName: "VerificationRequested",
          logs: receipt.logs.filter((log) => log.address.toLowerCase() === target.address.toLowerCase()),
        });
        if (!logs.length) throw new Error("The confirmed transaction did not emit a verification request.");
        setVerification(undefined);
        setPollError("");
        setRequestId(logs[0].args.requestId);
      }
      const message = operation === "requestVerification" ? "Verification requested." :
        operation === "authorizeOracle" ? "Oracle authorized." : "Oracle permission revoked.";
      setState({ phase: "confirmed", message, hash });
    } catch (error) {
      setState({ phase: "failed", message: errorMessage(error), hash });
    }
  }

  function authorize(event: FormEvent) {
    event.preventDefault();
    void send("authorizeOracle");
  }

  return (
    <section className="panel" aria-labelledby="wallet-title">
      <div className="row between">
        <h2 id="wallet-title">Ethereum wallet</h2>
        {account
          ? <Button variant="secondary" disabled={busy} onClick={() => { setAccount(undefined); setProvider(undefined); }}>Disconnect</Button>
          : <Button disabled={!target} onClick={connect}>Connect MetaMask</Button>}
      </div>
      {!target && <p>Configure the registry address for this network to enable wallet transactions.</p>}
      {account && <>
        <p className="hash">{account}</p>
        <form className="wallet-actions" onSubmit={authorize}>
          <label>Oracle address<input value={oracle} onChange={(event) => setOracle(event.target.value.trim())} placeholder="0x…" required /></label>
          <Button disabled={busy || !oracle}>Authorize</Button>
          <Button type="button" variant="secondary" disabled={busy || !oracle} onClick={() => send("revokeOracle")}>Revoke</Button>
        </form>
        <Button variant="secondary" disabled={busy || !recordId} onClick={() => send("requestVerification")}>Request on-chain verification</Button>
      </>}
      {account && target && badge && <Badges key={`${network}:${account}`} target={target} address={badge} account={account} />}
      {state && <p className={state.phase === "failed" ? "notice" : ""} role={state.phase === "failed" ? "alert" : "status"}>
        {state.message}{state.hash && <><br /><span className="hash">{state.hash}</span></>}
      </p>}
      {requestId !== undefined && <div aria-live="polite">
        <p>Request #{requestId.toString()} · {verification?.status || "Awaiting oracle"}</p>
        {verification?.stale && <p className="notice">The oracle flagged the scan as stale.</p>}
        {verification && <p>Fulfilled by <span className="hash">{verification.oracle}</span></p>}
        {pollError && <p>{pollError}</p>}
      </div>}
    </section>
  );
}
