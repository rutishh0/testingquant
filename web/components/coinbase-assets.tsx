"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { apiClient, type CoinbaseExchangeRatesResponse } from "@/components/api-client";
import { TrendingUp, RefreshCw } from "lucide-react";

export default function CoinbaseAssets() {
  const [exchangeRates, setExchangeRates] = useState<CoinbaseExchangeRatesResponse['data'] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = async () => {
    try {
      setLoading(true);
      setError(null);

      const ratesResponse = await apiClient.coinbase.getExchangeRates('USD');
      setExchangeRates(ratesResponse.data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch exchange rates');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center py-8">
        <RefreshCw className="w-6 h-6 animate-spin text-green-600" />
        <span className="ml-2 text-sm text-muted-foreground">Loading exchange rates...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-8">
        <div className="text-sm text-red-600 mb-2">{error}</div>
        <Button variant="outline" size="sm" onClick={fetchData}>
          Try Again
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-2">
          <TrendingUp className="w-4 h-4" />
          <span className="font-medium">Exchange Rates</span>
        </div>
        <Button variant="outline" size="sm" onClick={fetchData}>
          <RefreshCw className="w-4 h-4" />
        </Button>
      </div>

      {/* Exchange Rates Summary */}
      {exchangeRates && (
        <div className="bg-gradient-to-r from-green-50 to-blue-50 p-4 rounded-lg border">
          <h3 className="font-medium mb-2">Rates (Base: {exchangeRates.currency})</h3>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-sm">
            {Object.entries(exchangeRates.rates)
              .slice(0, 12)
              .map(([currency, rate]) => (
                <div key={currency} className="flex justify-between">
                  <span className="font-medium">{currency}:</span>
                  <span className="font-mono">{typeof rate === 'string' ? parseFloat(rate).toFixed(4) : rate}</span>
                </div>
              ))}
          </div>
          <div className="text-xs text-muted-foreground mt-2">
            Updated: {new Date(exchangeRates.updated_at).toLocaleString()}
          </div>
        </div>
      )}
    </div>
  );
}