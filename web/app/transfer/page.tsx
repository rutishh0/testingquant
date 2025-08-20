"use client";

import { useState } from "react";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Wallet } from "lucide-react";
import { JsonRpcProvider, Wallet as EthersWallet, parseEther, isAddress } from "ethers";

// Hardcoded Infura Sepolia endpoint (replace YOUR_INFURA_PROJECT_ID with your real key)
const DEFAULT_INFURA_SEPOLIA = "https://sepolia.infura.io/v3/YOUR_INFURA_PROJECT_ID";

export default function TransferPage() {
  const [rpcUrl, setRpcUrl] = useState(DEFAULT_INFURA_SEPOLIA);
  const [fromPrivKey, setFromPrivKey] = useState("");
  const [toAddress, setToAddress] = useState("");
  const [amountEth, setAmountEth] = useState("");
  const [sending, setSending] = useState(false);
  const [txHash, setTxHash] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function onSend(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setTxHash(null);

    try {
      if (!rpcUrl || rpcUrl.includes("YOUR_INFURA_PROJECT_ID")) {
        throw new Error("Please set your actual Infura Project ID in the hardcoded RPC URL.");
      }
      if (!fromPrivKey || !fromPrivKey.startsWith("0x") || fromPrivKey.length < 64) {
        throw new Error("Enter a valid sender private key (0x-prefixed)");
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
      const wallet = new EthersWallet(fromPrivKey, provider);

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

  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800 p-6">
      <div className="container mx-auto max-w-2xl">
        <div className="mb-6 flex items-center justify-between">
          <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-blue-600 to-violet-600 bg-clip-text text-transparent flex items-center gap-2">
            <Wallet className="w-7 h-7" /> Sepolia ETH Transfer
          </h1>
          <Badge variant="outline">Sepolia Testnet</Badge>
        </div>

        <Card className="border-blue-200 bg-blue-50/50 dark:border-blue-800 dark:bg-blue-950/50">
          <CardHeader>
            <CardTitle>Send ETH</CardTitle>
            <CardDescription>Use an EOA private key to sign locally and broadcast via Infura Sepolia</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={onSend} className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">RPC URL</label>
                <Input value={rpcUrl} onChange={(e) => setRpcUrl(e.target.value)} placeholder="Sepolia RPC URL" />
                <p className="text-xs text-muted-foreground mt-1">Hardcoded in this page for demonstration. Replace YOUR_INFURA_PROJECT_ID with your real key.</p>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">From Private Key</label>
                <Input type="password" value={fromPrivKey} onChange={(e) => setFromPrivKey(e.target.value)} placeholder="0x..." />
                <p className="text-xs text-muted-foreground mt-1">Your key never leaves the browser. The transaction is signed locally.</p>
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
                  <a className="text-sm underline" href={`https://sepolia.etherscan.io/tx/${txHash}`} target="_blank" rel="noreferrer">
                    View on Etherscan
                  </a>
                )}
              </div>

              {error && <p className="text-sm text-red-600">{error}</p>}
            </form>
          </CardContent>
        </Card>
      </div>
    </main>
  );
}