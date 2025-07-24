const API_BASE_URL = typeof window !== 'undefined' ? window.location.origin : 'http://localhost:8080';

async function post<T, U>(endpoint: string, body: T): Promise<U> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };

  // Add API key for authentication
  headers['X-API-Key'] = 'development-api-key-12345678901234567890';

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
  });

  if (!response.ok) {
    const errorBody = await response.text();
    throw new Error(`API Error: ${response.status} ${response.statusText} - ${errorBody}`);
  }

  return response.json();
}

export const apiClient = {
  post,
};
