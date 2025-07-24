"use client";

import { useState } from 'react';
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { apiClient } from './api-client';

// Define interfaces for the API request and response
interface CombineRequest {
  network_identifier: {
    blockchain: string;
    network: string;
  };
  unsigned_transaction: string;
  signatures: any[]; // Keeping signatures as any[] for flexibility
}

interface CombineResponse {
  signed_transaction: string;
}

export default function Combine() {
  const [requestBody, setRequestBody] = useState(JSON.stringify({
    network_identifier: { blockchain: "ethereum", network: "goerli" },
    unsigned_transaction: "",
    signatures: []
  }, null, 2));
  const [response, setResponse] = useState<CombineResponse | null>(null);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setResponse(null);
    try {
      const result = await apiClient.post<CombineRequest, CombineResponse>('/construction/combine', JSON.parse(requestBody));
      setResponse(result);
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('An unknown error occurred');
      }
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <Card>
        <CardHeader>
          <CardTitle>Request Body</CardTitle>
        </CardHeader>
        <CardContent>
          <Textarea
            className="min-h-[200px] font-mono"
            value={requestBody}
            onChange={(e) => setRequestBody(e.target.value)}
          />
        </CardContent>
        <CardFooter className="flex justify-between">
          <Button type="submit">Combine</Button>
        </CardFooter>
      </Card>
      {response && (
        <Card className="mt-4">
          <CardHeader>
            <CardTitle>Response</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="bg-gray-100 dark:bg-gray-800 p-4 rounded-md">
              {JSON.stringify(response, null, 2)}
            </pre>
          </CardContent>
        </Card>
      )}
      {error && (
        <Card className="mt-4">
          <CardHeader>
            <CardTitle className="text-red-500">Error</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-red-500">{error}</p>
          </CardContent>
        </Card>
      )}
    </form>
  );
}
