"use client";

import React, { useState } from 'react';
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { apiClient } from './api-client';

// Define interfaces for the API request and response
interface SubmitRequest {
  network_identifier: {
    blockchain: string;
    network: string;
  };
  signed_transaction: string;
}

interface SubmitResponse {
  transaction_identifier: {
    hash: string;
  };
  metadata: object;
}

export default function Submit() {
  const [requestBody, setRequestBody] = useState(JSON.stringify({
    network_identifier: { blockchain: "ethereum", network: "goerli" },
    signed_transaction: ""
  }, null, 2));
  const [response, setResponse] = useState<SubmitResponse | null>(null);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setResponse(null);
    try {
      const result = await apiClient.post<SubmitRequest, SubmitResponse>('/v1/construction/submit', JSON.parse(requestBody));
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
          <Button type="submit">Submit</Button>
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
