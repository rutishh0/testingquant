const API_BASE_URL = typeof window !== 'undefined' ? window.location.origin : 'http://localhost:8080';

// API response types
export interface HealthResponse {
  status: string;
  timestamp: number;
  services: Record<string, ServiceHealth>;
}

export interface ServiceHealth {
  status: string;
  error?: string;
}

export interface CoinbaseWallet {
  id: string;
  name: string;
  primary_address: string;
  default_network: {
    id: string;
    display_name: string;
    chain_id: number;
    is_testnet: boolean;
  };
  created_at: string;
  updated_at: string;
  features: string[];
}

export interface CoinbaseAsset {
  asset_id: string;
  name: string;
  symbol: string;
  decimals: number;
  display_name: string;
  address_format: string;
  explorer_url?: string;
  contract_address?: string;
  image_url?: string;
}

export interface OverledgerNetwork {
  id: string;
  name: string;
  description: string;
  type: string;
  status: string;
}

export interface CoinbaseExchangeRatesResponse {
  data: {
    currency: string;
    rates: Record<string, string>;
    updated_at: string;
  };
}

// Automated test runner result type (mirrors backend)
export interface TestResult {
  tier: number;
  name: string;
  success: boolean;
  message?: string;
  error?: string;
}

export interface CoinbaseExchangeRatesResponse {
  data: {
    currency: string;
    rates: Record<string, string>;
    updated_at: string;
  };
}

// Utility function for API calls
async function apiCall<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...((options.headers as Record<string, string>) || {}),
  };

  // Add API key for authentication if provided at build/runtime
  const apiKey = process.env.NEXT_PUBLIC_API_KEY;
  if (apiKey) {
    headers['X-API-Key'] = apiKey;
  }

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    let errorMessage = `API Error: ${response.status} ${response.statusText}`;
    try {
      const errorBody = await response.json();
      if (errorBody.message) {
        errorMessage = errorBody.message;
      }
    } catch {
      // Use default error message if parsing fails
    }
    throw new Error(errorMessage);
  }

  return response.json();
}

// Health & Status
const health = {
  check: (): Promise<HealthResponse> => apiCall('/health'),
};

// Coinbase API methods
const coinbase = {
  // Wallets
  getWallets: () => apiCall<{ data: CoinbaseWallet[] }>('/v1/coinbase/wallets'),
  
  createWallet: (data: { name: string; use_server_signer?: boolean }) =>
    apiCall<{ data: CoinbaseWallet }>('/v1/coinbase/wallets', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  getWalletBalance: (walletId: string) =>
    apiCall(`/v1/coinbase/wallets/${walletId}/balance`),

  // Assets & Networks
  getAssets: () => apiCall<{ data: CoinbaseAsset[] }>('/v1/coinbase/assets'),
  
  getNetworks: () => apiCall('/v1/coinbase/networks'),
  
  getExchangeRates: (currency?: string): Promise<CoinbaseExchangeRatesResponse> => {
    const query = currency ? `?currency=${currency}` : '';
    return apiCall(`/v1/coinbase/exchange-rates${query}`);
  },

  // Transactions
  createTransaction: (walletId: string, data: {
    amount: string;
    asset_id: string;
    destination: string;
    network?: string;
    speed?: string;
  }) =>
    apiCall(`/v1/coinbase/wallets/${walletId}/transactions`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  getTransaction: (transactionId: string) =>
    apiCall(`/v1/coinbase/transactions/${transactionId}`),

  estimateFee: (walletId: string, data: {
    amount: string;
    asset_id: string;
    destination: string;
    network_id?: string;
    speed?: string;
  }) =>
    apiCall(`/v1/coinbase/wallets/${walletId}/transactions/estimate-fee`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),
};

// Overledger API methods
const overledger = {
  getNetworks: () => apiCall<{ networks: OverledgerNetwork[] }>('/v1/overledger/networks'),
  
  getBalance: (networkId: string, address: string) =>
    apiCall(`/v1/overledger/networks/${networkId}/addresses/${address}/balance`),

  createTransaction: (data: {
    networkId: string;
    fromAddress: string;
    toAddress: string;
    amount: string;
    tokenId?: string;
    gasLimit?: string;
    gasPrice?: string;
  }) =>
    apiCall('/v1/overledger/transactions', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  getTransactionStatus: (networkId: string, txHash: string) =>
    apiCall(`/v1/overledger/networks/${networkId}/transactions/${txHash}/status`),

  testConnection: () => apiCall('/v1/overledger/test'),
};

// Exchange API methods
const exchange = {
  listProducts: () => apiCall('/v1/exchange/products'),
};

// Test runner API
const tests = {
  getResults: () => apiCall<TestResult[]>('/tests'),
};

export const apiClient = {
  health,
  exchange,
  coinbase,
  overledger,
  mesh: {
    getNetworks: () =>
      apiCall<{
        networks: Array<{
          network_identifier: { blockchain: string; network: string };
          currency?: { symbol: string; decimals: number };
        }>;
      }>('/v1/mesh/networks'),
    getAccountBalance: (data: {
      network_identifier: unknown;
      account_identifier: unknown;
    }) =>
      apiCall<{
        balances: Array<{ value: string; currency: { symbol: string; decimals: number } }>;
      }>('/v1/mesh/account/balance', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  },
  tests,
};
