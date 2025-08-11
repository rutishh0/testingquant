"use client";

import React, { useEffect, useRef, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { apiClient } from "@/components/api-client";

interface TickerMessage {
  type: string;
  product_id: string;
  price: string;
  open_24h: string;
  volume_24h: string;
  low_24h: string;
  high_24h: string;
  time: string;
}

interface ProductSummary {
  productId?: string;
  product_id?: string;
  id?: string;
}

export default function ExchangeMarketData() {
  const [selectedProduct, setSelectedProduct] = useState<string>("BTC-USD");
  const [products, setProducts] = useState<string[]>([]);
  const [ticker, setTicker] = useState<TickerMessage | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  // Fetch available products once
  useEffect(() => {
    apiClient.exchange
      .listProducts()
      .then((resp) => {
        const cast = resp as { products?: ProductSummary[]; data?: ProductSummary[] };
        const items = cast.products ?? cast.data ?? [];
        const ids = items
          .map((p: ProductSummary) => p.productId || p.product_id || p.id)
          .filter((id): id is string => typeof id === "string");
        setProducts(ids.sort());
      })
      .catch((err) => console.error("Failed to fetch products", err));
  }, []);

  // Manage WebSocket connection when selected product changes
  useEffect(() => {
    // Close previous ws if exists
    if (wsRef.current) {
      wsRef.current.close();
    }

    const ws = new WebSocket("wss://ws-feed.exchange.coinbase.com");
    wsRef.current = ws;

    ws.onopen = () => {
      const sub = {
        type: "subscribe",
        product_ids: [selectedProduct],
        channels: ["ticker"],
      };
      ws.send(JSON.stringify(sub));
    };

    ws.onmessage = (evt) => {
      try {
        const data: TickerMessage = JSON.parse(evt.data);
        if (data.type === "ticker") {
          setTicker(data);
        }
      } catch {
        // ignore parse errors
      }
    };

    ws.onerror = (err) => console.error("WebSocket error", err);

    return () => {
      ws.close();
    };
  }, [selectedProduct]);

  const price = ticker ? parseFloat(ticker.price) : undefined;
  const open = ticker ? parseFloat(ticker.open_24h) : undefined;
  const pctChange = price && open ? ((price - open) / open) * 100 : 0;

  return (
    <Card className="border-green-200 bg-green-50/40 dark:border-green-800 dark:bg-green-900/20">
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <span>Coinbase Exchange – Live Ticker</span>
          <select
            className="rounded-md border bg-transparent p-1 text-sm dark:border-slate-700"
            value={selectedProduct}
            onChange={(e: React.ChangeEvent<HTMLSelectElement>) => setSelectedProduct(e.target.value)}
          >
            {products.map((p: string) => (
              <option key={p} value={p}>
                {p}
              </option>
            ))}
          </select>
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {ticker ? (
          <div className="text-center space-y-3">
            <div className="text-4xl font-semibold">
              ${price?.toLocaleString(undefined, {
                minimumFractionDigits: 2,
                maximumFractionDigits: 2,
              })}
            </div>
            <Badge variant={pctChange >= 0 ? "default" : "destructive"}>
              {pctChange >= 0 ? "+" : ""}
              {pctChange.toFixed(2)}%
            </Badge>
            <div className="text-xs text-muted-foreground">
              24h Vol: {parseFloat(ticker.volume_24h).toLocaleString()}
            </div>
          </div>
        ) : (
          <div className="text-center text-sm text-muted-foreground">Connecting…</div>
        )}
      </CardContent>
    </Card>
  );
}
