"use client";

import React, { useEffect, useRef, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { apiClient } from "@/components/api-client";
import { AlertCircle } from "lucide-react";

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

const FALLBACK_PRODUCTS = ["BTC-USD", "ETH-USD", "SOL-USD", "ADA-USD", "DOGE-USD"];

export default function ExchangeMarketData() {
  const [selectedProduct, setSelectedProduct] = useState<string>("BTC-USD");
  const [products, setProducts] = useState<string[]>([]);
  const [ticker, setTicker] = useState<TickerMessage | null>(null);
  const [warning, setWarning] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const mountedRef = useRef<boolean>(false);

  // Fetch available products once
  useEffect(() => {
    mountedRef.current = true;
    apiClient.exchange
      .listProducts()
      .then((resp) => {
        const cast = resp as { products?: ProductSummary[]; data?: ProductSummary[] };
        const items = cast.products ?? cast.data ?? [];
        const ids = items
          .map((p: ProductSummary) => p.productId || p.product_id || p.id)
          .filter((id): id is string => typeof id === "string");

        if (ids.length === 0) {
          // Backend returned no products (likely not configured). Fallback to a sensible default list.
          setWarning(
            "Exchange API not configured or returned no products. Showing default markets."
          );
          setProducts(FALLBACK_PRODUCTS);
          if (!FALLBACK_PRODUCTS.includes(selectedProduct)) {
            setSelectedProduct(FALLBACK_PRODUCTS[0]);
          }
          return;
        }

        const sorted = ids.sort();
        setProducts(sorted);
        // Ensure current selection is valid
        if (!sorted.includes(selectedProduct)) {
          setSelectedProduct(sorted[0]);
        }
      })
      .catch((err) => {
        console.warn("Failed to fetch products", err);
        setWarning(
          err instanceof Error
            ? `${err.message}. Showing default markets.`
            : "Exchange API unavailable. Showing default markets."
        );
        setProducts(FALLBACK_PRODUCTS);
        if (!FALLBACK_PRODUCTS.includes(selectedProduct)) {
          setSelectedProduct(FALLBACK_PRODUCTS[0]);
        }
      });
    return () => {
      mountedRef.current = false;
    };
  }, [selectedProduct]);

  // Manage WebSocket connection when selected product changes
  useEffect(() => {
    // Close previous ws if exists
    if (wsRef.current) {
      try { wsRef.current.close(); } catch {}
    }

    // Guard: WebSocket may not be available in some SSR/test contexts
    if (typeof window === 'undefined' || typeof WebSocket === 'undefined') {
      return;
    }

    const ws = new WebSocket("wss://ws-feed.exchange.coinbase.com");
    wsRef.current = ws;

    ws.onopen = () => {
      try {
        const sub = {
          type: "subscribe",
          product_ids: [selectedProduct],
          channels: ["ticker"],
        };
        ws.send(JSON.stringify(sub));
      } catch (e) {
        // ignore send errors
      }
    };

    ws.onmessage = (evt) => {
      try {
        const data: TickerMessage = JSON.parse(evt.data as string);
        if (data && data.type === "ticker") {
          setTicker(data);
        }
      } catch {
        // ignore parse errors
      }
    };

    ws.onerror = () => {
      // Avoid noisy error objects in the console; surface via UI if needed
      if (mountedRef.current && !warning) {
        setWarning(
          "Live ticker connection encountered an error. Data may be delayed."
        );
      }
    };

    ws.onclose = () => {
      // No-op; component will recreate a connection on next selection change
    };

    return () => {
      try { ws.close(); } catch {}
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
            disabled={products.length === 0}
          >
            {products.length === 0 ? (
              <option>Loading…</option>
            ) : (
              products.map((p: string) => (
                <option key={p} value={p}>
                  {p}
                </option>
              ))
            )}
          </select>
        </CardTitle>
        {warning && (
          <div className="mt-2 flex items-center gap-2 text-xs text-yellow-700">
            <AlertCircle className="h-4 w-4" />
            <span>{warning}</span>
          </div>
        )}
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
              24h Vol: {ticker?.volume_24h ? parseFloat(ticker.volume_24h).toLocaleString() : "-"}
            </div>
          </div>
        ) : (
          <div className="text-center text-sm text-muted-foreground">Connecting…</div>
        )}
      </CardContent>
    </Card>
  );
}
