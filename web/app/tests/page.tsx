"use client";

import { useEffect, useMemo, useState } from "react";
import { apiClient, type TestResult } from "@/components/api-client";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { RefreshCw, ClipboardCopy, ShieldCheck, ShieldX } from "lucide-react";

export default function TestsPage() {
  const [results, setResults] = useState<TestResult[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const fetchResults = async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await apiClient.tests.getResults();
      setResults(res);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to fetch test results");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchResults();
    const id = setInterval(fetchResults, 30000);
    return () => clearInterval(id);
  }, []);

  const grouped = useMemo(() => {
    const g: Record<number, TestResult[]> = {};
    (Array.isArray(results) ? results : []).forEach((r) => {
      g[r.tier] = g[r.tier] ? [...g[r.tier], r] : [r];
    });
    return g;
  }, [results]);

  const plainText = useMemo(() => {
    if (!Array.isArray(results)) return "";
    const lines: string[] = [];
    const byTier: Record<number, TestResult[]> = {};
    results.forEach((r) => {
      byTier[r.tier] = byTier[r.tier] ? [...byTier[r.tier], r] : [r];
    });
    Object.keys(byTier)
      .map((k) => Number(k))
      .sort((a, b) => a - b)
      .forEach((tier) => {
        lines.push(`Tier ${tier}`);
        byTier[tier].forEach((r) => {
          const status = r.success ? "PASS" : "FAIL";
          const parts = [
            `- ${r.name}: ${status}`,
            r.message ? `  message: ${r.message}` : "",
            r.error ? `  error: ${r.error}` : "",
          ].filter(Boolean);
          lines.push(parts.join("\n"));
        });
        lines.push("");
      });
    return lines.join("\n");
  }, [results]);

  const copyAll = async () => {
    try {
      await navigator.clipboard.writeText(plainText || "");
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {}
  };

  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
      <div className="container mx-auto px-6 py-8 space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold">Automated Test Results</h1>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" onClick={fetchResults}>
              <RefreshCw className="w-4 h-4" />
            </Button>
            <Button variant="outline" size="sm" onClick={copyAll}>
              <ClipboardCopy className="w-4 h-4 mr-1" />
              {copied ? "Copied" : "Copy All"}
            </Button>
          </div>
        </div>

        {loading && (
          <div className="flex items-center text-sm text-muted-foreground">
            <RefreshCw className="w-4 h-4 mr-2 animate-spin" /> Loading...
          </div>
        )}
        {error && (
          <div className="text-sm text-red-600">{error}</div>
        )}

        {/* Detailed list */}
        {Array.isArray(results) && (
          <div className="space-y-6">
            {Object.keys(grouped)
              .map((k) => Number(k))
              .sort((a, b) => a - b)
              .map((tier) => (
                <div key={tier} className="border rounded-lg p-4 bg-white/60 dark:bg-slate-900/40">
                  <h2 className="font-semibold mb-3">Tier {tier}</h2>
                  <div className="space-y-3">
                    {grouped[tier].map((r, idx) => (
                      <div key={idx} className="border rounded p-3 bg-white dark:bg-slate-900">
                        <div className="flex items-center justify-between">
                          <div className="font-medium">{r.name}</div>
                          {r.success ? (
                            <Badge className="bg-green-600/10 text-green-700 dark:bg-green-800/30 dark:text-green-300">
                              <ShieldCheck className="w-4 h-4 mr-1" /> Pass
                            </Badge>
                          ) : (
                            <Badge className="bg-red-600/10 text-red-700 dark:bg-red-800/30 dark:text-red-300">
                              <ShieldX className="w-4 h-4 mr-1" /> Fail
                            </Badge>
                          )}
                        </div>
                        {r.message && (
                          <div className="text-xs mt-2 text-muted-foreground">{r.message}</div>
                        )}
                        {r.error && (
                          <div className="text-xs mt-2 text-red-600 whitespace-pre-wrap">{r.error}</div>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              ))}
          </div>
        )}

        {/* Raw copyable block */}
        {Array.isArray(results) && (
          <div className="mt-8">
            <h2 className="font-semibold mb-2">Copyable Text</h2>
            <textarea
              className="w-full h-56 text-xs p-3 border rounded bg-white dark:bg-slate-900"
              readOnly
              value={plainText}
            />
          </div>
        )}
      </div>
    </main>
  );
}


