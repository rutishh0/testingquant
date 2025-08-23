"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Wallet } from "lucide-react";
import { JsonRpcProvider, Wallet as EthersWallet, parseEther, isAddress } from "ethers";
import { apiClient } from "@/components/api-client";

// Default to public env var if provided; otherwise fall back to the user's provided project id
const DEFAULT_INFURA_SEPOLIA = (process.env.NEXT_PUBLIC_INFURA_RPC_URL as string) || "https://sepolia.infura.io/v3/fffa2b97b0ee4febb791505fd5c067de";

type Direction = "eoa" | "overledger";

export default function TransferPage() {
  // Shared UI state
  const [direction, setDirection] = useState<Direction>("eoa");

  // EOA (local signing) state
  const [rpcUrl, setRpcUrl] = useState(DEFAULT_INFURA_SEPOLIA);
  const [fromPrivKey, setFromPrivKey] = useState("");
  const [derivedFromAddress, setDerivedFromAddress] = useState<string | null>(null);
  const [toAddress, setToAddress] = useState("");
  const [amountEth, setAmountEth] = useState("");
  const [sending, setSending] = useState(false);
  const [txHash, setTxHash] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Overledger (server-side) state
  const [olNetworks, setOlNetworks] = useState<Array<{ id: string; name: string }>>([]);
  const [selectedNetworkId, setSelectedNetworkId] = useState<string>("");
  const [olFromAddress, setOlFromAddress] = useState("");
  const [olToAddress, setOlToAddress] = useState("");
  const [olAmount, setOlAmount] = useState("");
  const [olSending, setOlSending] = useState(false);
  const [olTxHash, setOlTxHash] = useState<string | null>(null);
  const [olStatus, setOlStatus] = useState<string | null>(null);
  const [olMessage, setOlMessage] = useState<string | null>(null);
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [advGas, setAdvGas] = useState("");
  const [advMaxFee, setAdvMaxFee] = useState("");
  const [advMaxPriority, setAdvMaxPriority] = useState("");
  const [advNonce, setAdvNonce] = useState<string>("");
  const pollRef = useRef<NodeJS.Timeout | null>(null);

  // Helpers
  function maybeAutoprefixKey(input: string): string {
    const trimmed = input.trim();
    const hex64 = /^[0-9a-fA-F]{64}$/;
    if (hex64.test(trimmed) && !trimmed.startsWith("0x")) {
      return "0x" + trimmed;
    }
    return input;
  }

  function updateDerivedAddress(pk: string) {
    try {
      // Only attempt when looks like a hex key 0x + 64 chars
      const looksValid = /^0x[0-9a-fA-F]{64}$/.test(pk.trim());
      if (!looksValid) {
        setDerivedFromAddress(null);
        return;
      }
      const w = new EthersWallet(pk);
      setDerivedFromAddress(w.address);
    } catch {
      setDerivedFromAddress(null);
    }
  }

  // Handlers
  async function onSend(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setTxHash(null);

    try {
      if (!rpcUrl) {
        throw new Error("Please set a valid RPC URL.");
      }
      const pk = fromPrivKey.trim();
      if (!/^0x[0-9a-fA-F]{64}$/.test(pk)) {
        throw new Error("Enter a valid sender private key (0x-prefixed hex, 64 chars)");
      }
      if (!isAddress(toAddress)) {
        throw new Error("Enter a valid destination address");
      }
      const value = amountEth.trim();
      if (!value || Number(value) <= 0) {
        throw new Error("Enter a valid positive amount in ETH");
      }

      setSending(true);
      const provider = new JsonRpcProvider(rpcUrl);
      const wallet = new EthersWallet(pk, provider);

      const tx = await wallet.sendTransaction({
        to: toAddress,
        value: parseEther(value),
      });

      const receipt = await tx.wait();
      setTxHash(receipt?.hash ?? tx.hash);
    } catch (err: any) {
      setError(err?.message ?? String(err));
    } finally {
      setSending(false);
    }
  }

  function onPrivKeyChange(value: string) {
    const maybe = maybeAutoprefixKey(value);
    setFromPrivKey(maybe);
    updateDerivedAddress(maybe);
  }

  // Overledger actions
  async function loadNetworksOnce() {
    if (olNetworks.length > 0) return;
    try {
      const res = await apiClient.overledger.getNetworks();
      const networks = (res.networks || []).map((n: any) => ({ id: n.id, name: n.name || n.id }));
      setOlNetworks(networks);
      if (networks.length > 0) setSelectedNetworkId(networks[0].id);
    } catch (e: any) {
      // Non-fatal for the EOA flow
      console.error("Failed to load Overledger networks", e?.message || e);
    }
  }

  useEffect(() => {
    if (direction === "overledger") {
      loadNetworksOnce();
    }
    // Cleanup any polling when switching away
    return () => {
      if (pollRef.current) {
        clearInterval(pollRef.current);
        pollRef.current = null;
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [direction]);

  // Simple heuristic to detect EVM networks from networkId
  function isEvmNetworkId(id: string): boolean {
    const s = (id || "").toLowerCase();
    return [
      "eth",
      "ethereum",
      "sepolia",
      "holesky",
      "goerli",
      "polygon",
      "matic",
      "bsc",
      "binance",
      "arbitrum",
      "optimism",
      "base",
      "avalanche",
      "celo",
      "fantom",
      "linea",
      "zksync",
      "scroll",
    ].some((k) => s.includes(k));
  }

  async function onSendOverledger(e: React.FormEvent) {
    e.preventDefault();
    setOlTxHash(null);
    setOlStatus(null);
    setOlMessage(null);

    if (!selectedNetworkId) {
      setOlStatus("Select a network");
      return;
    }
    if (!isAddress(olToAddress) || !isAddress(olFromAddress)) {
      setOlStatus("Enter valid from/to addresses");
      return;
    }
    const value = olAmount.trim();
    if (!value || Number(value) <= 0) {
      setOlStatus("Enter a valid positive amount");
      return;
    }

    try {
      setOlSending(true);

      // Send decimal ETH amount as-is to match Overledger curl format
      const amountForApi = value;

      const overrides: any = {};
      if (advGas.trim()) overrides.gasLimit = advGas.trim();
      if (advMaxFee.trim()) overrides.maxFeePerGas = advMaxFee.trim();
      if (advMaxPriority.trim()) overrides.maxPriorityFeePerGas = advMaxPriority.trim();
      if (advNonce.trim()) {
        const n = parseInt(advNonce.trim(), 10);
        if (!Number.isNaN(n)) (overrides as any).nonce = n;
      }

      const resp = await apiClient.overledger.createTransaction({
        networkId: selectedNetworkId,
        fromAddress: olFromAddress,
        toAddress: olToAddress,
        amount: amountForApi,
        ...overrides,
      });
      // Expect TransactionResponse shape
      const meta = (resp as any)?.metadata || {};
      const txHashResp =
        (resp as any).transactionId ||
        (resp as any).hash ||
        meta.transactionId ||
        null;
      const execMsg = (resp as any)?.metadata?.message || (resp as any)?.message;
      const execInfo = (resp as any)?.metadata?.execution;
      if (txHashResp) setOlTxHash(txHashResp);
      const initialStatus = (resp as any).status || execInfo?.value || "created";
      setOlStatus(initialStatus);
      if (execMsg) setOlMessage(execMsg);
      if (!execMsg && execInfo?.description) setOlMessage(execInfo.description);

      // Start simple polling of status for a short window
      if (txHashResp) {
        let attempts = 0;
        pollRef.current = setInterval(async () => {
          attempts += 1;
          try {
            const status: any = await apiClient.overledger.getTransactionStatus(selectedNetworkId, txHashResp);
            const s = (status as any)?.status || (status as any)?.value || "pending";
            setOlStatus(s);
            const sLower = String(s).toLowerCase();
            // capture tx id from status
            const tid = (status as any)?.transactionId || (status as any)?.hash;
            if (tid) setOlTxHash(tid);
            if (["confirmed", "failed"].includes(sLower)) {
              if (pollRef.current) clearInterval(pollRef.current);
              pollRef.current = null;
            }
          } catch (e) {
            // ignore transient errors
          }
          if (attempts > 20) {
            if (pollRef.current) clearInterval(pollRef.current);
            pollRef.current = null;
          }
        }, 3000);
      }
    } catch (e: any) {
      setOlStatus(e?.message || "Failed to create transaction");
    } finally {
      setOlSending(false);
    }
  }

  // Explorer helper
  const explorerTxUrl = useMemo(() => {
    if (!txHash) return null;
    // Default to Sepolia
    return `https://sepolia.etherscan.io/tx/${txHash}`;
  }, [txHash]);

  const explorerTxUrlOL = useMemo(() => {
    if (!olTxHash) return null;
    // Best-effort: if network mentions Sepolia use etherscan; otherwise generic
    const lower = (selectedNetworkId || "").toLowerCase();
    if (lower.includes("sepolia") || lower.includes("ethereum")) {
      return `https://sepolia.etherscan.io/tx/${olTxHash}`;
    }
    return null;
  }, [olTxHash, selectedNetworkId]);

  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800 p-6">
      <div className="container mx-auto max-w-2xl">
        <div className="mb-6 flex items-center justify-between">
          <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-blue-600 to-violet-600 bg-clip-text text-transparent flex items-center gap-2">
            <Wallet className="w-7 h-7" /> Sepolia ETH Transfer
          </h1>
          <Badge variant="outline">Sepolia Testnet</Badge>
        </div>

        {/* Direction toggle */}
        <div className="mb-4 flex gap-2">
          <Button variant={direction === "eoa" ? "default" : "secondary"} onClick={() => setDirection("eoa")}>EOA → Address</Button>
          <Button variant={direction === "overledger" ? "default" : "secondary"} onClick={() => setDirection("overledger")}>Overledger → Address</Button>
        </div>

        {/* EOA Transfer Card */}
        {direction === "eoa" && (
          <Card className="border-blue-200 bg-blue-50/50 dark:border-blue-800 dark:bg-blue-950/50">
            <CardHeader>
              <CardTitle>Send ETH (Local Signing)</CardTitle>
              <CardDescription>Use an EOA private key to sign locally and broadcast via Infura Sepolia</CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={onSend} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium mb-1">RPC URL</label>
                  <Input value={rpcUrl} onChange={(e) => setRpcUrl(e.target.value)} placeholder="Sepolia RPC URL" />
                  <p className="text-xs text-muted-foreground mt-1">Uses your NEXT_PUBLIC_INFURA_RPC_URL if set; otherwise defaults to your provided project ID.</p>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">From Private Key</label>
                  <Input
                    type="password"
                    value={fromPrivKey}
                    onChange={(e) => onPrivKeyChange(e.target.value)}
                    placeholder="0x..."
                  />
                  {derivedFromAddress && (
                    <p className="text-xs mt-1">Derived sender address: <span className="font-mono">{derivedFromAddress}</span></p>
                  )}
                  <p className="text-xs text-muted-foreground mt-1">We auto-add 0x when you paste a 64-character hex key. Your key never leaves the browser.</p>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">To Address</label>
                  <Input value={toAddress} onChange={(e) => setToAddress(e.target.value)} placeholder="0xRecipient" />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Amount (ETH)</label>
                  <Input value={amountEth} onChange={(e) => setAmountEth(e.target.value)} placeholder="0.01" />
                </div>

                <div className="flex items-center gap-3">
                  <Button type="submit" disabled={sending}>
                    {sending ? "Sending..." : "Send ETH"}
                  </Button>
                  {txHash && (
                    <a className="text-sm underline" href={explorerTxUrl ?? undefined} target="_blank" rel="noreferrer">
                      View on Etherscan
                    </a>
                  )}
                </div>

                {error && <p className="text-sm text-red-600">{error}</p>}
              </form>
            </CardContent>
          </Card>
        )}

        {/* Overledger Transfer Card */}
        {direction === "overledger" && (
          <Card className="border-violet-200 bg-violet-50/50 dark:border-violet-800 dark:bg-violet-950/50">
            <CardHeader>
              <CardTitle>Send via Overledger</CardTitle>
              <CardDescription>Create a server-side transaction using Overledger APIs</CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={onSendOverledger} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium mb-1">Network</label>
                  <select
                    className="w-full border rounded-md h-10 px-3 bg-white dark:bg-slate-900"
                    value={selectedNetworkId}
                    onChange={(e) => setSelectedNetworkId(e.target.value)}
                  >
                    {olNetworks.length === 0 && <option value="">Loading networks...</option>}
                    {olNetworks.map((n) => (
                      <option key={n.id} value={n.id}>{n.name || n.id}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">From Address</label>
                  <Input value={olFromAddress} onChange={(e) => setOlFromAddress(e.target.value)} placeholder="0xSender" />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">To Address</label>
                  <Input value={olToAddress} onChange={(e) => setOlToAddress(e.target.value)} placeholder="0xRecipient" />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Amount</label>
                  <Input value={olAmount} onChange={(e) => setOlAmount(e.target.value)} placeholder="0.01" />
                  <p className="text-xs text-muted-foreground mt-1">
                    Enter amount in ETH (decimal). This value is sent verbatim to Overledger.
                  </p>
                </div>

                <div className="flex items-center gap-3 flex-wrap">
                  <Button type="submit" disabled={olSending || !selectedNetworkId}>
                    {olSending ? "Submitting..." : "Submit Transaction"}
                  </Button>
                  {olTxHash && explorerTxUrlOL && (
                    <a href={explorerTxUrlOL} target="_blank" rel="noreferrer">
                      <Button type="button" variant="outline">View on Etherscan</Button>
                    </a>
                  )}
                </div>

                {/* Advanced overrides */}
                <div className="mt-3">
                  <Button type="button" variant="secondary" onClick={() => setShowAdvanced((s) => !s)}>
                    {showAdvanced ? "Hide Advanced" : "Show Advanced"}
                  </Button>
                </div>
                {showAdvanced && (
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mt-3">
                    <div>
                      <label className="block text-sm font-medium mb-1">Gas</label>
                      <Input value={advGas} onChange={(e) => setAdvGas(e.target.value)} placeholder="22086" />
                    </div>
                    <div>
                      <label className="block text-sm font-medium mb-1">Max Fee Per Gas (wei)</label>
                      <Input value={advMaxFee} onChange={(e) => setAdvMaxFee(e.target.value)} placeholder="9618390" />
                    </div>
                    <div>
                      <label className="block text-sm font-medium mb-1">Max Priority Fee Per Gas (wei)</label>
                      <Input value={advMaxPriority} onChange={(e) => setAdvMaxPriority(e.target.value)} placeholder="1500000" />
                    </div>
                    <div>
                      <label className="block text-sm font-medium mb-1">Nonce</label>
                      <Input value={advNonce} onChange={(e) => setAdvNonce(e.target.value)} placeholder="1" />
                    </div>
                  </div>
                )}

                {olStatus && (
                  <p className="text-sm">
                    Status: {olStatus}
                    {explorerTxUrlOL && (
                      <>
                        {" "}—{" "}
                        <a className="underline" href={explorerTxUrlOL} target="_blank" rel="noreferrer">
                          View on Etherscan
                        </a>
                      </>
                    )}
                  </p>
                )}
                {olMessage && <p className="text-sm text-muted-foreground break-words">Message: {olMessage}</p>}
                {/* Show code if available in metadata */}
                {(() => {
                  const m: any = (olTxHash ? undefined : undefined) as any; // placeholder to keep type relaxed
                  return null;
                })()}
              </form>
            </CardContent>
          </Card>
        )}
      </div>
    </main>
  );
}