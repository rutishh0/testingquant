"use client";

import { useEffect, useState } from "react";
import { apiClient } from "@/components/api-client";
import { Card } from "@/components/ui/card";
import { RefreshCw } from "lucide-react";

type MeshNetwork = {
  network_identifier: { blockchain: string; network: string };
  currency?: { symbol: string; decimals: number };
};

export default function MeshNetworks() {
  const [networks, setNetworks] = useState<MeshNetwork[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await apiClient.mesh.getNetworks();
      setNetworks(res.networks || []);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load networks");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  if (loading) {
    return (
      <div className="flex items-center space-x-2 text-sm text-muted-foreground">
        <RefreshCw className="w-4 h-4 animate-spin" />
        <span>Loading Mesh networks...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-sm text-red-600">
        {error}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 gap-3">
      {networks.map((n: MeshNetwork, idx: number) => (
        <Card key={idx} className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="font-medium">
                {n.network_identifier.blockchain} / {n.network_identifier.network}
              </div>
              {n.currency && (
                <div className="text-xs text-muted-foreground">
                  Currency: {n.currency.symbol} (decimals {n.currency.decimals})
                </div>
              )}
            </div>
          </div>
        </Card>
      ))}
    </div>
  );
}


