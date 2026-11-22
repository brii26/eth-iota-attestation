import { useEffect, useRef, useState } from "react";
import type { FormEvent } from "react";
import { history, networks, statusText, verify } from "./api";
import type { History, Selection, Verification } from "./api";
import { compareWithRegistry } from "./ethereum";
import { Button } from "./components/ui/button";
import { Wallet } from "./Wallet";

const networkNames: Record<string, string> = {
  "ethereum-private": "Private Ethereum",
  "ethereum-sepolia": "Ethereum Sepolia",
  "iota-testnet": "IOTA Testnet",
  "iota-localnet": "IOTA Localnet",
};
const pageSize = 10;
const blank: History = { total: 0, records: [] };
const date = (value: string) => new Date(value).toLocaleString();

function age(value: string) {
  const hours = Math.max(0, Math.floor((Date.now() - Date.parse(value)) / 3_600_000));
  return hours < 1 ? "Less than an hour" : hours < 48 ? `${hours} h` : `${Math.floor(hours / 24)} d`;
}

function Result({ result, comparison }: { result: Verification; comparison: string }) {
  const status = statusText[result.status];
  return (
    <section className="panel result" aria-labelledby="result-title" aria-live="polite">
      <div className="row between">
        <h2 id="result-title">Verification result</h2>
        <span className={`badge ${result.status === "VALID" ? "valid" : "warning"}`}>{status.label}</span>
      </div>
      <p>{status.explanation}</p>
      {result.stale && <p role="status" className="notice">The Grype database is stale or its build time is unavailable.</p>}
      <dl className="details">
        <div><dt>Package</dt><dd>{result.checks.artifact}</dd></div>
        <div><dt>SBOM</dt><dd>{result.checks.sbom}</dd></div>
        <div><dt>Scan</dt><dd>{result.checks.scan}</dd></div>
        {result.signer && <div><dt>Recording oracle</dt><dd className="hash">{result.signer}</dd></div>}
        {result.recordedAt && <div><dt>Recorded</dt><dd>{date(result.recordedAt)} · {age(result.recordedAt)} old</dd></div>}
        {result.grypeDbBuiltAt && <div><dt>Grype database build</dt><dd>{date(result.grypeDbBuiltAt)}</dd></div>}
      </dl>
      {comparison && <p className={comparison.startsWith("Matches") ? "match" : "notice"}>{comparison}</p>}
    </section>
  );
}

export function App() {
  const [networkList, setNetworkList] = useState<string[]>([]);
  const [draft, setDraft] = useState<Selection>({ network: "", app: "", version: "" });
  const [selection, setSelection] = useState<Selection>();
  const [data, setData] = useState<History>(blank);
  const [page, setPage] = useState(0);
  const [selectedId, setSelectedId] = useState<string>();
  const [result, setResult] = useState<Verification>();
  const [comparison, setComparison] = useState("");
  const [loading, setLoading] = useState(false);
  const [verifying, setVerifying] = useState(false);
  const [networkError, setNetworkError] = useState("");
  const [historyError, setHistoryError] = useState("");
  const [verifyError, setVerifyError] = useState("");
  const verifyRequest = useRef<AbortController>(null);

  useEffect(() => {
    const controller = new AbortController();
    networks(controller.signal).then(({ networks: available }) => {
      setNetworkList(available);
      setDraft((current) => ({ ...current, network: current.network || available[0] || "" }));
    }).catch((error) => { if (!controller.signal.aborted) setNetworkError(error.message); });
    return () => {
      controller.abort();
      verifyRequest.current?.abort();
    };
  }, []);

  useEffect(() => {
    if (!selection) return;
    const controller = new AbortController();
    setLoading(true);
    setHistoryError("");
    history(selection, page * pageSize, pageSize, controller.signal).then(setData)
      .catch((error) => { if (!controller.signal.aborted) setHistoryError(error.message); })
      .finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => controller.abort();
  }, [selection, page]);

  function chooseRecord(id: string | undefined) {
    verifyRequest.current?.abort();
    setSelectedId(id);
    setResult(undefined);
    setComparison("");
    setVerifyError("");
    setVerifying(false);
  }

  function find(event: FormEvent) {
    event.preventDefault();
    chooseRecord(undefined);
    setData(blank);
    setPage(0);
    setSelection({ network: draft.network, app: draft.app.trim(), version: draft.version.trim() });
  }

  async function runVerification() {
    if (!selection) return;
    verifyRequest.current?.abort();
    const controller = new AbortController();
    verifyRequest.current = controller;
    setVerifying(true);
    setVerifyError("");
    setResult(undefined);
    setComparison("");
    try {
      const next = await verify(selection, selectedId, controller.signal);
      if (controller.signal.aborted) return;
      setResult(next);
      if (selection.network.startsWith("ethereum-") && next.recordHashes) {
        setComparison("Reading the Ethereum registry…");
        try {
          const matches = await compareWithRegistry(selection, selectedId, next);
          if (!controller.signal.aborted) setComparison(matches
            ? "Matches the record read directly from the Ethereum registry."
            : "The direct Ethereum registry read differs from the oracle result.");
        } catch (error) {
          if (!controller.signal.aborted) setComparison(`Direct Ethereum read unavailable. ${(error as Error).message}`);
        }
      }
    } catch (error) {
      if (!controller.signal.aborted) setVerifyError((error as Error).message);
    } finally {
      if (!controller.signal.aborted) setVerifying(false);
    }
  }

  function changeNetwork(network: string) {
    setDraft({ ...draft, network });
    if (selection) {
      chooseRecord(undefined);
      setData(blank);
      setPage(0);
      setSelection({ ...selection, network });
    }
  }

  return (
    <main className="shell">
      <header className="page-header">
        <h1>Artifact Attestation</h1>
        <p>Look up a software release and check it against its ledger record.</p>
      </header>
      <section className="panel" aria-labelledby="search-title">
        <h2 id="search-title">Find a release</h2>
        {networkError && <p role="alert" className="notice">{networkError}</p>}
        {!networkError && networkList.length === 0 && <p role="status">Waiting for configured networks.</p>}
        <form className="search" onSubmit={find}>
          <label>Network<select value={draft.network} onChange={(event) => changeNetwork(event.target.value)} required>
            {!networkList.length && <option value="">Select a network</option>}
            {networkList.map((network) => <option key={network} value={network}>{networkNames[network] || network}</option>)}
          </select></label>
          <label>Package<input value={draft.app} onChange={(event) => setDraft({ ...draft, app: event.target.value })} placeholder="@scope/package" required /></label>
          <label>Version<input value={draft.version} onChange={(event) => setDraft({ ...draft, version: event.target.value })} placeholder="1.0.0" required /></label>
          <Button disabled={!draft.network || !draft.app.trim() || !draft.version.trim()}>Find release</Button>
        </form>
      </section>
      {selection && <section className="panel" aria-labelledby="history-title">
        <div className="row between">
          <div><h2 id="history-title">Record history</h2><p>{selection.app}@{selection.version} · {networkNames[selection.network] || selection.network}</p></div>
          <Button disabled={verifying} onClick={runVerification}>{verifying ? "Verifying…" : selectedId ? "Verify selected record" : "Verify latest"}</Button>
        </div>
        {loading && <p role="status">Loading record history…</p>}
        {historyError && <p role="alert" className="notice">{historyError}</p>}
        {verifyError && <p role="alert" className="notice">{verifyError}</p>}
        {!loading && !historyError && !data.total && <p>No records in the index for this release.</p>}
        {!loading && data.total > 0 && <div className="scroll">
          <table>
            <thead><tr><th scope="col">Select</th><th scope="col">Scan ID</th><th scope="col">Status</th><th scope="col">Submitted</th><th scope="col">Transaction</th></tr></thead>
            <tbody>{data.records.map((item) => <tr key={item.scanId}>
              <td><input type="radio" name="record" aria-label={`Select scan ${item.scanId}`} checked={selectedId === item.recordId && Boolean(item.recordId)} disabled={!item.recordId || item.status !== "included"} onChange={() => chooseRecord(item.recordId)} /></td>
              <td className="hash">{item.scanId}</td>
              <td><span className="badge">{item.status}</span></td>
              <td>{date(item.submittedAt)}</td>
              <td className="hash">{item.txRef || "Pending reference"}</td>
            </tr>)}</tbody>
          </table>
        </div>}
        <div className="row controls">
          {selectedId && <Button variant="secondary" onClick={() => chooseRecord(undefined)}>Use latest</Button>}
          {data.total > pageSize && <>
            <Button variant="secondary" disabled={page === 0 || loading} onClick={() => setPage(page - 1)}>Previous</Button>
            <span>Page {page + 1} of {Math.ceil(data.total / pageSize)}</span>
            <Button variant="secondary" disabled={(page + 1) * pageSize >= data.total || loading} onClick={() => setPage(page + 1)}>Next</Button>
          </>}
        </div>
      </section>}
      {result && <Result result={result} comparison={comparison} />}
      {draft.network.startsWith("ethereum-") && <Wallet key={draft.network} network={draft.network}
        recordId={selection?.network === draft.network ? selectedId || result?.recordId : undefined} />}
    </main>
  );
}
