import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { 
  Wallet, 
  Network, 
  TrendingUp, 
  Activity,
  ExternalLink,
  Shield,
  ShieldCheck,
  Zap
} from "lucide-react";
import CoinbaseWallets from "@/components/coinbase-wallets";
import CoinbaseAssets from "@/components/coinbase-assets";
import OverledgerNetworks from "@/components/overledger-networks";
import OverledgerTransactions from "@/components/overledger-transactions";
import SystemHealth from "@/components/system-health";
import TestingLogs from "@/components/testing-logs";

export default function Home() {
  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
      {/* Header */}
      <div className="border-b bg-white/50 backdrop-blur-sm dark:bg-slate-900/50">
        <div className="container mx-auto px-6 py-8">
          <div className="flex items-center justify-between">
            <div className="space-y-2">
              <h1 className="text-4xl font-bold tracking-tight bg-gradient-to-r from-blue-600 to-violet-600 bg-clip-text text-transparent">
                Quant Connector
              </h1>
              <p className="text-lg text-muted-foreground">
                Production-ready Coinbase and Overledger integration platform
              </p>
            </div>
            <div className="flex items-center space-x-4">
              <Badge variant="outline" className="px-3 py-1">
                <Shield className="w-4 h-4 mr-2" />
                Production Ready
              </Badge>
              <Badge variant="outline" className="px-3 py-1">
                <Zap className="w-4 h-4 mr-2" />
                Real-time
              </Badge>
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="container mx-auto px-6 py-8">
        {/* System Health Overview */}
        <div className="mb-8">
          <SystemHealth />
        </div>

        {/* Main Tabs */}
        <Tabs defaultValue="coinbase" className="space-y-6">
          <TabsList className="grid w-full grid-cols-5 h-12">
            <TabsTrigger value="coinbase" className="flex items-center space-x-2">
              <Wallet className="w-4 h-4" />
              <span>Coinbase</span>
            </TabsTrigger>
            <TabsTrigger value="overledger" className="flex items-center space-x-2">
              <Network className="w-4 h-4" />
              <span>Overledger</span>
            </TabsTrigger>
            <TabsTrigger value="trading" className="flex items-center space-x-2">
              <TrendingUp className="w-4 h-4" />
              <span>Trading</span>
            </TabsTrigger>
            <TabsTrigger value="testing" className="flex items-center space-x-2">
              <ShieldCheck className="w-4 h-4" />
              <span>Testing</span>
            </TabsTrigger>
            <TabsTrigger value="analytics" className="flex items-center space-x-2">
              <Activity className="w-4 h-4" />
              <span>Analytics</span>
            </TabsTrigger>
          </TabsList>

          {/* Coinbase Tab */}
          <TabsContent value="coinbase" className="space-y-6">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <Card className="border-blue-200 bg-blue-50/50 dark:border-blue-800 dark:bg-blue-950/50">
                <CardHeader>
                  <CardTitle className="flex items-center space-x-2">
                    <Wallet className="w-5 h-5 text-blue-600" />
                    <span>Wallet Management</span>
                  </CardTitle>
                  <CardDescription>
                    Create, manage, and monitor your Coinbase wallets and addresses
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <CoinbaseWallets />
                </CardContent>
              </Card>

              <Card className="border-green-200 bg-green-50/50 dark:border-green-800 dark:bg-green-950/50">
              <CardHeader>
                  <CardTitle className="flex items-center space-x-2">
                    <TrendingUp className="w-5 h-5 text-green-600" />
                    <span>Assets & Trading</span>
                  </CardTitle>
                <CardDescription>
                    View available assets, exchange rates, and trading pairs
                </CardDescription>
              </CardHeader>
              <CardContent>
                  <CoinbaseAssets />
              </CardContent>
            </Card>
            </div>
          </TabsContent>

          {/* Overledger Tab */}
          <TabsContent value="overledger" className="space-y-6">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <Card className="border-purple-200 bg-purple-50/50 dark:border-purple-800 dark:bg-purple-950/50">
                <CardHeader>
                  <CardTitle className="flex items-center space-x-2">
                    <Network className="w-5 h-5 text-purple-600" />
                    <span>Network Explorer</span>
                  </CardTitle>
                  <CardDescription>
                    Explore supported blockchain networks and their capabilities
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <OverledgerNetworks />
                </CardContent>
              </Card>

              <Card className="border-orange-200 bg-orange-50/50 dark:border-orange-800 dark:bg-orange-950/50">
              <CardHeader>
                  <CardTitle className="flex items-center space-x-2">
                    <Activity className="w-5 h-5 text-orange-600" />
                    <span>Cross-Chain Transactions</span>
                  </CardTitle>
                <CardDescription>
                    Execute and monitor cross-chain transactions across networks
                </CardDescription>
              </CardHeader>
              <CardContent>
                  <OverledgerTransactions />
              </CardContent>
            </Card>
            </div>
          </TabsContent>

          {/* Testing Logs Tab */}
          <TabsContent value="testing" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <ShieldCheck className="w-5 h-5" />
                  <span>Testing Logs</span>
                </CardTitle>
                <CardDescription>
                  Automated backend test results (refreshes every 30s)
                </CardDescription>
              </CardHeader>
              <CardContent>
                <TestingLogs />
              </CardContent>
            </Card>
          </TabsContent>

          {/* Trading Tab */}
          <TabsContent value="trading" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <TrendingUp className="w-5 h-5" />
                  <span>Trading Dashboard</span>
                </CardTitle>
                <CardDescription>
                  Advanced trading features and portfolio management
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="text-center py-12 text-muted-foreground">
                  <TrendingUp className="w-16 h-16 mx-auto mb-4 opacity-50" />
                  <h3 className="text-lg font-semibold">Trading Features Coming Soon</h3>
                  <p>Advanced trading and portfolio management tools will be available in the next release.</p>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* Analytics Tab */}
          <TabsContent value="analytics" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <Activity className="w-5 h-5" />
                  <span>Analytics Dashboard</span>
                </CardTitle>
                <CardDescription>
                  Transaction analytics, performance metrics, and insights
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="text-center py-12 text-muted-foreground">
                  <Activity className="w-16 h-16 mx-auto mb-4 opacity-50" />
                  <h3 className="text-lg font-semibold">Analytics Coming Soon</h3>
                  <p>Comprehensive analytics and reporting features will be available in the next release.</p>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>

        {/* Footer */}
        <div className="mt-16 pt-8 border-t text-center text-sm text-muted-foreground">
          <p>
            Powered by{" "}
            <a
              href="https://coinbase.com"
              target="_blank"
              rel="noopener noreferrer"
              className="text-blue-600 hover:text-blue-700 inline-flex items-center"
            >
              Coinbase <ExternalLink className="w-3 h-3 ml-1" />
            </a>{" "}
            and{" "}
            <a
              href="https://quant.network"
              target="_blank"
              rel="noopener noreferrer"
              className="text-purple-600 hover:text-purple-700 inline-flex items-center"
            >
              Quant Overledger <ExternalLink className="w-3 h-3 ml-1" />
            </a>
          </p>
        </div>
      </div>
    </main>
  );
}

