"use client";

import { Activity } from "lucide-react";

export default function OverledgerTransactions() {
  return (
    <div className="text-center py-8 text-muted-foreground">
      <Activity className="w-12 h-12 mx-auto mb-4 opacity-50" />
      <h3 className="text-lg font-semibold">Cross-Chain Transactions</h3>
      <p className="text-sm">Cross-chain transaction functionality coming soon.</p>
    </div>
  );
} 