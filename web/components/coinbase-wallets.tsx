"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { apiClient, type CoinbaseWallet } from "@/components/api-client";
import { Wallet, Plus, RefreshCw, Copy, ExternalLink } from "lucide-react";

export default function CoinbaseWallets() {
  const [wallets, setWallets] = useState<CoinbaseWallet[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [newWalletName, setNewWalletName] = useState('');
  const [showCreateDialog, setShowCreateDialog] = useState(false);

  const fetchWallets = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await apiClient.coinbase.getWallets();
      setWallets(response.data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch wallets');
    } finally {
      setLoading(false);
    }
  };

  const createWallet = async () => {
    if (!newWalletName.trim()) return;
    
    try {
      setCreating(true);
      const response = await apiClient.coinbase.createWallet({
        name: newWalletName.trim(),
        use_server_signer: true,
      });
      setWallets([...wallets, response.data]);
      setNewWalletName('');
      setShowCreateDialog(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create wallet');
    } finally {
      setCreating(false);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  useEffect(() => {
    fetchWallets();
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center py-8">
        <RefreshCw className="w-6 h-6 animate-spin text-blue-600" />
        <span className="ml-2 text-sm text-muted-foreground">Loading wallets...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-8">
        <div className="text-sm text-red-600 mb-2">{error}</div>
        <Button variant="outline" size="sm" onClick={fetchWallets}>
          Try Again
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-2">
          <Wallet className="w-4 h-4" />
          <span className="font-medium">Wallets ({wallets.length})</span>
        </div>
        <div className="flex items-center space-x-2">
          <Button variant="outline" size="sm" onClick={fetchWallets}>
            <RefreshCw className="w-4 h-4" />
          </Button>
          <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
            <DialogTrigger asChild>
              <Button size="sm">
                <Plus className="w-4 h-4" />
                New Wallet
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Create New Wallet</DialogTitle>
              </DialogHeader>
              <div className="space-y-4">
                <div>
                  <label className="text-sm font-medium">Wallet Name</label>
                  <Input
                    value={newWalletName}
                    onChange={(e) => setNewWalletName(e.target.value)}
                    placeholder="Enter wallet name"
                    className="mt-1"
                  />
                </div>
                <div className="flex justify-end space-x-2">
                  <Button
                    variant="outline"
                    onClick={() => setShowCreateDialog(false)}
                  >
                    Cancel
                  </Button>
                  <Button
                    onClick={createWallet}
                    disabled={creating || !newWalletName.trim()}
                  >
                    {creating ? (
                      <>
                        <RefreshCw className="w-4 h-4 animate-spin mr-2" />
                        Creating...
                      </>
                    ) : (
                      'Create Wallet'
                    )}
                  </Button>
                </div>
              </div>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {wallets.length === 0 ? (
        <div className="text-center py-8 text-muted-foreground">
          <Wallet className="w-12 h-12 mx-auto mb-4 opacity-50" />
          <p>No wallets found</p>
          <p className="text-sm">Create your first wallet to get started</p>
        </div>
      ) : (
        <div className="space-y-3">
          {wallets.map((wallet) => (
            <div
              key={wallet.id}
              className="border rounded-lg p-4 hover:shadow-sm transition-shadow"
            >
              <div className="flex items-start justify-between">
                <div className="space-y-2 flex-1">
                  <div className="flex items-center space-x-2">
                    <h3 className="font-medium">{wallet.name}</h3>
                    <Badge variant="outline" className="text-xs">
                      {wallet.default_network.display_name}
                    </Badge>
                    {wallet.default_network.is_testnet && (
                      <Badge variant="outline" className="text-xs bg-yellow-50 text-yellow-700 border-yellow-200">
                        Testnet
                      </Badge>
                    )}
                  </div>
                  
                  <div className="flex items-center space-x-2 text-sm text-muted-foreground">
                    <span className="font-mono text-xs">
                      {wallet.primary_address.slice(0, 8)}...{wallet.primary_address.slice(-6)}
                    </span>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-6 w-6 p-0"
                      onClick={() => copyToClipboard(wallet.primary_address)}
                    >
                      <Copy className="w-3 h-3" />
                    </Button>
                  </div>

                  <div className="text-xs text-muted-foreground">
                    Created: {new Date(wallet.created_at).toLocaleDateString()}
                  </div>
                </div>
                
                <div className="flex items-center space-x-2">
                  <Button variant="outline" size="sm">
                    View Balance
                  </Button>
                  <Button variant="ghost" size="sm">
                    <ExternalLink className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
} 