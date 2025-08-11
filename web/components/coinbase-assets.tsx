"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { apiClient, type CoinbaseAsset, type CoinbaseExchangeRatesResponse } from "@/components/api-client";
import { TrendingUp, RefreshCw, ExternalLink } from "lucide-react";

export default function CoinbaseAssets() {
  const [assets, setAssets] = useState<CoinbaseAsset[]>([]);
  const [exchangeRates, setExchangeRates] = useState<CoinbaseExchangeRatesResponse['data'] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const [assetsResponse, ratesResponse] = await Promise.all([
        apiClient.coinbase.getAssets(),
        apiClient.coinbase.getExchangeRates('USD')
      ]);
      
      setAssets(assetsResponse.data);
      setExchangeRates(ratesResponse.data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch assets');
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
        <span className="ml-2 text-sm text-muted-foreground">Loading assets...</span>
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

  const popularAssets = assets.slice(0, 6); // Show first 6 assets

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-2">
          <TrendingUp className="w-4 h-4" />
          <span className="font-medium">Assets & Rates</span>
        </div>
        <Button variant="outline" size="sm" onClick={fetchData}>
          <RefreshCw className="w-4 h-4" />
        </Button>
      </div>

      {/* Exchange Rates Summary */}
      {exchangeRates && (
        <div className="bg-gradient-to-r from-green-50 to-blue-50 p-4 rounded-lg border">
          <h3 className="font-medium mb-2">Exchange Rates (USD)</h3>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-sm">
            {Object.entries(exchangeRates.rates)
              .slice(0, 8)
              .map(([currency, rate]) => (
                <div key={currency} className="flex justify-between">
                  <span className="font-medium">{currency}:</span>
                  <span className="font-mono">${typeof rate === 'string' ? parseFloat(rate).toFixed(2) : rate}</span>
                </div>
              ))}
          </div>
          <div className="text-xs text-muted-foreground mt-2">
            Updated: {new Date(exchangeRates.updated_at).toLocaleString()}
          </div>
        </div>
      )}

      {/* Popular Assets */}
      <div>
        <h3 className="font-medium mb-3">Popular Assets</h3>
        {popularAssets.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            <TrendingUp className="w-12 h-12 mx-auto mb-4 opacity-50" />
            <p>No assets available</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-3">
            {popularAssets.map((asset) => (
              <div
                key={asset.asset_id}
                className="border rounded-lg p-4 hover:shadow-sm transition-shadow"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    {asset.image_url && (
                      <img
                        src={asset.image_url}
                        alt={asset.name}
                        className="w-8 h-8 rounded-full"
                        onError={(e) => {
                          (e.target as HTMLImageElement).style.display = 'none';
                        }}
                      />
                    )}
                    <div>
                      <div className="flex items-center space-x-2">
                        <h4 className="font-medium">{asset.symbol}</h4>
                        <Badge variant="outline" className="text-xs">
                          {asset.decimals} decimals
                        </Badge>
                      </div>
                      <p className="text-sm text-muted-foreground">{asset.display_name}</p>
                    </div>
                  </div>
                  
                  <div className="flex items-center space-x-2">
                    {exchangeRates?.rates[asset.symbol] && (
                      <div className="text-right">
                        <div className="font-mono text-sm">
                          ${parseFloat(exchangeRates.rates[asset.symbol]).toFixed(2)}
                        </div>
                      </div>
                    )}
                    {asset.explorer_url && (
                      <Button variant="ghost" size="sm" asChild>
                        <a href={asset.explorer_url} target="_blank" rel="noopener noreferrer">
                          <ExternalLink className="w-4 h-4" />
                        </a>
                      </Button>
                    )}
                  </div>
                </div>
                
                {asset.contract_address && (
                  <div className="mt-2 text-xs text-muted-foreground">
                    Contract: <span className="font-mono">{asset.contract_address.slice(0, 8)}...{asset.contract_address.slice(-6)}</span>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
        
        {assets.length > 6 && (
          <div className="text-center mt-4">
            <Button variant="outline" size="sm">
              View All Assets ({assets.length})
            </Button>
          </div>
        )}
      </div>
    </div>
  );
} 