"use client";

import { useEffect, useMemo, useState, useCallback } from "react";
import { apiClient, type TestResult } from "@/components/api-client";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { RefreshCw, ClipboardCopy, ShieldCheck, ShieldX, Search, MinusCircle } from "lucide-react";

export default function TestsPage() {
  const [results, setResults] = useState<TestResult[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [query, setQuery] = useState("");

  // Helpers for status/pretty printing
  const isSkip = useCallback((r: TestResult) => {
    const t = `${r.message || ""} ${r.error || ""}`;
    return /(^|\b)skip(?:ped)?(?::|\b)/i.test(t);
  }, []);

  const statusOf = useCallback((r: TestResult) => {
    if (isSkip(r) && r.success) return "skip" as const;
    return r.success ? "pass" : "fail";
  }, [isSkip]);

  const prettyText = useCallback((text?: string) => {
    if (!text) return "";
    const trimmed = text.trim();
    try {
      if ((trimmed.startsWith("{") && trimmed.endsWith("}")) || (trimmed.startsWith("[") && trimmed.endsWith("]"))) {
        const obj = JSON.parse(trimmed);
        return JSON.stringify(obj, null, 2);
      }
    } catch {}
    return trimmed
      .replace(/\r\n/g, "\n")
      .replace(/ \-\-\- /g, " \n--- ")
      .replace(/ === /g, " \n=== ")
      .replace(/ -> /g, " \n-> ");
  }, []);

  const fetchResults = useCallback(async () => {
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
  }, []);

  useEffect(() => {
    fetchResults();
    const id = setInterval(fetchResults, 30000);
    return () => clearInterval(id);
  }, [fetchResults]);

  // All results array (safety coercion)
  const allResults = useMemo(() => (Array.isArray(results) ? results : []), [results]);

  // Stats for header
  const stats = useMemo(() => {
    const total = allResults.length;
    const passed = allResults.filter((r) => statusOf(r) === "pass").length;
    const failed = allResults.filter((r) => statusOf(r) === "fail").length;
    const passRate = total ? Math.round((passed / total) * 100) : 0;
    return { total, passed, failed, passRate };
  }, [allResults, statusOf]);

  // Filtered lists
  const filteredAll = useMemo(() => {
    if (!query.trim()) return allResults;
    const q = query.toLowerCase();
    return allResults.filter((r) =>
      [r.name, r.message || "", r.error || ""].some((s) => s.toLowerCase().includes(q))
    );
  }, [allResults, query]);

  const filteredFails = useMemo(() => filteredAll.filter((r) => statusOf(r) === "fail"), [filteredAll, statusOf]);

  // Group by tier helper
  const groupByTier = useCallback((list: TestResult[]) => {
    const g: Record<number, TestResult[]> = {};
    list.forEach((r) => {
      g[r.tier] = g[r.tier] ? [...g[r.tier], r] : [r];
    });
    return g;
  }, []);

  const groupedAll = useMemo(() => groupByTier(filteredAll), [filteredAll, groupByTier]);
  const groupedFails = useMemo(() => groupByTier(filteredFails), [filteredFails, groupByTier]);

  const plainText = useMemo(() => {
    if (!Array.isArray(allResults)) return "";
    const lines: string[] = [];
    const byTier: Record<number, TestResult[]> = {};
    allResults.forEach((r) => {
      byTier[r.tier] = byTier[r.tier] ? [...byTier[r.tier], r] : [r];
    });
    Object.keys(byTier)
      .map((k) => Number(k))
      .sort((a, b) => a - b)
      .forEach((tier) => {
        lines.push(`Tier ${tier}`);
        byTier[tier].forEach((r) => {
          const st = statusOf(r) === "pass" ? "PASS" : statusOf(r) === "skip" ? "SKIP" : "FAIL";
          const parts = [
            `- ${r.name}: ${st}`,
            r.message ? `  message: ${prettyText(r.message)}` : "",
            r.error ? `  error: ${prettyText(r.error)}` : "",
          ].filter(Boolean);
          lines.push(parts.join("\n"));
        });
        lines.push("");
      });
    return lines.join("\n");
  }, [allResults, prettyText, statusOf]);

  const copyAll = async () => {
    try {
      await navigator.clipboard.writeText(plainText || "");
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {}
  };

  // Loading & error states
  if (loading) {
    return (
      <main className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
        <div className="container mx-auto px-6 py-10">
          <div className="flex items-center text-sm text-muted-foreground">
            <RefreshCw className="w-4 h-4 mr-2 animate-spin" /> Loading test results...
          </div>
        </div>
      </main>
    );
  }

  if (error) {
    return (
      <main className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
        <div className="container mx-auto px-6 py-10 space-y-4">
          <div className="text-sm text-red-600">{error}</div>
          <Button variant="outline" size="sm" onClick={fetchResults}>
            <RefreshCw className="w-4 h-4 mr-1" /> Try Again
          </Button>
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
      <div className="container mx-auto px-6 py-8 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">Automated Test Results</h1>
            <p className="text-sm text-muted-foreground mt-1">Live results from conformance, integration, and validation suites</p>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" onClick={fetchResults}>
              <RefreshCw className="w-4 h-4 mr-1" /> Refresh
            </Button>
          </div>
        </div>

        {/* Summary Cards */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card>
            <CardHeader>
              <CardTitle>Total</CardTitle>
              <CardDescription>All checks</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-semibold">{stats.total}</div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Passed</CardTitle>
              <CardDescription>Successful checks</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center">
                <Badge className="bg-green-600/10 text-green-700 dark:bg-green-800/30 dark:text-green-300">
                  <ShieldCheck className="w-4 h-4 mr-1" /> {stats.passed}
                </Badge>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Failed</CardTitle>
              <CardDescription>Checks to review</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center">
                <Badge className="bg-red-600/10 text-red-700 dark:bg-red-800/30 dark:text-red-300">
                  <ShieldX className="w-4 h-4 mr-1" /> {stats.failed}
                </Badge>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Pass Rate</CardTitle>
              <CardDescription>Overall success</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-semibold">{stats.passRate}%</div>
            </CardContent>
          </Card>
        </div>

        {/* Search / controls */}
        <div className="flex flex-col sm:flex-row gap-3 items-stretch sm:items-center justify-between">
          <div className="relative w-full sm:max-w-sm">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <Input
              placeholder="Search tests, messages, or errors..."
              className="pl-9"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
            />
          </div>
        </div>

        {/* Tabs */}
        <Tabs defaultValue="overview">
          <TabsList>
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="failures">Failures</TabsTrigger>
            <TabsTrigger value="raw">Raw</TabsTrigger>
          </TabsList>

          {/* Overview Tab */}
          <TabsContent value="overview" className="mt-4">
            {filteredAll.length === 0 ? (
              <div className="text-sm text-muted-foreground">No matching tests.</div>
            ) : (
              <div className="space-y-6">
                {Object.keys(groupedAll)
                  .map((k) => Number(k))
                  .sort((a, b) => a - b)
                  .map((tier) => (
                    <Card key={tier}>
                      <CardHeader>
                        <CardTitle>Tier {tier}</CardTitle>
                        <CardDescription>
                          {groupedAll[tier].filter((r) => statusOf(r) === "pass").length} passed · {groupedAll[tier].filter((r) => statusOf(r) === "fail").length} failed
                        </CardDescription>
                      </CardHeader>
                      <CardContent>
                        <div className="space-y-3">
                          {groupedAll[tier].map((r, idx) => {
                            const status = statusOf(r);
                            return (
                              <div key={idx} className="border rounded p-3 bg-white dark:bg-slate-900">
                                <div className="flex items-center justify-between">
                                  <div className="font-medium">{r.name}</div>
                                  {status === "pass" ? (
                                    <Badge className="bg-green-600/10 text-green-700 dark:bg-green-800/30 dark:text-green-300">
                                      <ShieldCheck className="w-4 h-4 mr-1" /> Pass
                                    </Badge>
                                  ) : status === "skip" ? (
                                    <Badge className="bg-amber-500/10 text-amber-700 dark:bg-amber-800/30 dark:text-amber-300">
                                      <MinusCircle className="w-4 h-4 mr-1" /> Skip
                                    </Badge>
                                  ) : (
                                    <Badge className="bg-red-600/10 text-red-700 dark:bg-red-800/30 dark:text-red-300">
                                      <ShieldX className="w-4 h-4 mr-1" /> Fail
                                    </Badge>
                                  )}
                                </div>
                                {r.message && (
                                  <pre className="text-xs mt-3 text-slate-700 dark:text-slate-300 whitespace-pre-wrap font-mono leading-relaxed bg-slate-50 dark:bg-slate-950/40 border rounded p-3 overflow-x-auto">{prettyText(r.message)}</pre>
                                )}
                                {r.error && (
                                  <pre className="text-xs mt-3 text-red-700 dark:text-red-300 whitespace-pre-wrap font-mono leading-relaxed bg-slate-50 dark:bg-slate-950/40 border rounded p-3 overflow-x-auto">{prettyText(r.error)}</pre>
                                )}
                              </div>
                            );
                          })}
                        </div>
                      </CardContent>
                    </Card>
                  ))}
              </div>
            )}
          </TabsContent>

          {/* Failures Tab */}
          <TabsContent value="failures" className="mt-4">
            {filteredFails.length === 0 ? (
              <div className="text-sm text-muted-foreground">No failures found.</div>
            ) : (
              <div className="space-y-6">
                {Object.keys(groupedFails)
                  .map((k) => Number(k))
                  .sort((a, b) => a - b)
                  .map((tier) => (
                    <Card key={tier}>
                      <CardHeader>
                        <CardTitle>Tier {tier}</CardTitle>
                        <CardDescription>
                          {groupedFails[tier].length} failing checks
                        </CardDescription>
                      </CardHeader>
                      <CardContent>
                        <div className="space-y-3">
                          {groupedFails[tier].map((r, idx) => (
                            <div key={idx} className="border rounded p-3 bg-white dark:bg-slate-900">
                              <div className="flex items-center justify-between">
                                <div className="font-medium">{r.name}</div>
                                <Badge className="bg-red-600/10 text-red-700 dark:bg-red-800/30 dark:text-red-300">
                                  <ShieldX className="w-4 h-4 mr-1" /> Fail
                                </Badge>
                              </div>
                              {r.message && (
                                <pre className="text-xs mt-3 text-slate-700 dark:text-slate-300 whitespace-pre-wrap font-mono leading-relaxed bg-slate-50 dark:bg-slate-950/40 border rounded p-3 overflow-x-auto">{prettyText(r.message)}</pre>
                              )}
                              {r.error && (
                                <pre className="text-xs mt-3 text-red-700 dark:text-red-300 whitespace-pre-wrap font-mono leading-relaxed bg-slate-50 dark:bg-slate-950/40 border rounded p-3 overflow-x-auto">{prettyText(r.error)}</pre>
                              )}
                            </div>
                          ))}
                        </div>
                      </CardContent>
                    </Card>
                  ))}
              </div>
            )}
          </TabsContent>

          {/* Raw Tab */}
          <TabsContent value="raw" className="mt-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <div>
                  <CardTitle>Raw Output</CardTitle>
                  <CardDescription>Plaintext, grouped by tier</CardDescription>
                </div>
                <Button variant="outline" size="sm" onClick={copyAll}>
                  <ClipboardCopy className="w-4 h-4 mr-1" /> {copied ? "Copied" : "Copy All"}
                </Button>
              </CardHeader>
              <CardContent>
                <pre className="w-full h-56 text-xs p-3 border rounded bg-white dark:bg-slate-900 overflow-auto whitespace-pre-wrap">{plainText}</pre>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </main>
  );
}


