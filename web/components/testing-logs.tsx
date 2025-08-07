"use client";

import { useEffect, useState } from "react";
import { apiClient, type TestResult } from "@/components/api-client";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { RefreshCw, ShieldCheck, ShieldX } from "lucide-react";

export default function TestingLogs() {
  const [results, setResults] = useState<TestResult[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

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

  // Fetch on mount and every 30s
  useEffect(() => {
    fetchResults();
    const id = setInterval(fetchResults, 30000);
    return () => clearInterval(id);
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center py-8">
        <RefreshCw className="w-6 h-6 animate-spin text-blue-600" />
        <span className="ml-2 text-sm text-muted-foreground">Running tests...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-8">
        <div className="text-sm text-red-600 mb-2">{error}</div>
        <Button variant="outline" size="sm" onClick={fetchResults}>
          Try Again
        </Button>
      </div>
    );
  }

  if (!results || results.length === 0) {
    return (
      <div className="text-center py-8 text-muted-foreground">No test results available.</div>
    );
  }

  // Group by tier
  const grouped: Record<number, TestResult[]> = {};
  results.forEach((r) => {
    grouped[r.tier] = grouped[r.tier] ? [...grouped[r.tier], r] : [r];
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h3 className="font-medium">Automated Testing Logs</h3>
        <Button variant="outline" size="sm" onClick={fetchResults}>
          <RefreshCw className="w-4 h-4" />
        </Button>
      </div>

      {Object.keys(grouped)
        .sort((a, b) => Number(a) - Number(b))
        .map((tierKey) => {
          const tier = Number(tierKey);
          return (
            <div key={tier} className="border rounded-lg p-4">
              <h4 className="font-medium mb-3">Tier {tier}</h4>
              <div className="space-y-2">
                {grouped[tier].map((r, idx) => (
                  <div key={idx} className="flex items-center justify-between">
                    <span>{r.name}</span>
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
                ))}
              </div>
            </div>
          );
        })}
    </div>
  );
}
