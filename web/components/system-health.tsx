"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { apiClient, type HealthResponse } from "@/components/api-client";
import { CheckCircle, XCircle, AlertCircle, RefreshCw } from "lucide-react";

export default function SystemHealth() {
  const [health, setHealth] = useState<HealthResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchHealth = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await apiClient.health.check();
      setHealth(response);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch health status');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchHealth();
    // Refresh health status every 30 seconds
    const interval = setInterval(fetchHealth, 30000);
    return () => clearInterval(interval);
  }, []);

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'healthy':
        return <CheckCircle className="w-4 h-4 text-green-500" />;
      case 'unhealthy':
        return <XCircle className="w-4 h-4 text-red-500" />;
      case 'degraded':
        return <AlertCircle className="w-4 h-4 text-yellow-500" />;
      case 'not_configured':
        return <AlertCircle className="w-4 h-4 text-gray-500" />;
      default:
        return <AlertCircle className="w-4 h-4 text-gray-500" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy':
        return 'bg-green-50 text-green-700 border-green-200';
      case 'unhealthy':
        return 'bg-red-50 text-red-700 border-red-200';
      case 'degraded':
        return 'bg-yellow-50 text-yellow-700 border-yellow-200';
      case 'not_configured':
        return 'bg-gray-50 text-gray-700 border-gray-200';
      default:
        return 'bg-gray-50 text-gray-700 border-gray-200';
    }
  };

  if (loading) {
    return (
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center space-x-2">
            <RefreshCw className="w-5 h-5 animate-spin" />
            <span>System Health</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-sm text-muted-foreground">Loading system status...</div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className="border-red-200 bg-red-50">
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center space-x-2 text-red-700">
            <XCircle className="w-5 h-5" />
            <span>System Health</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-sm text-red-600">{error}</div>
          <button
            onClick={fetchHealth}
            className="mt-2 text-sm text-red-600 hover:text-red-700 underline"
          >
            Try again
          </button>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            {getStatusIcon(health?.status || 'unknown')}
            <span>System Health</span>
          </div>
          <Badge
            variant="outline"
            className={getStatusColor(health?.status || 'unknown')}
          >
            {health?.status?.toUpperCase() || 'UNKNOWN'}
          </Badge>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {health?.services && Object.entries(health.services).map(([serviceName, serviceHealth]) => (
            <div
              key={serviceName}
              className="flex items-center justify-between p-3 rounded-lg border bg-white/50"
            >
              <div className="flex items-center space-x-3">
                {getStatusIcon(serviceHealth.status)}
                <div>
                  <div className="font-medium capitalize">{serviceName}</div>
                  {serviceHealth.error && (
                    <div className="text-xs text-red-600 mt-1">{serviceHealth.error}</div>
                  )}
                </div>
              </div>
              <Badge
                variant="outline"
                className={`text-xs ${getStatusColor(serviceHealth.status)}`}
              >
                {serviceHealth.status}
              </Badge>
            </div>
          ))}
        </div>
        
        {health?.timestamp && (
          <div className="mt-4 pt-3 border-t text-xs text-muted-foreground">
            Last updated: {new Date(health.timestamp * 1000).toLocaleString()}
          </div>
        )}
      </CardContent>
    </Card>
  );
} 