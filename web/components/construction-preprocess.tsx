"use client";

import { useState } from 'react';
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { apiClient } from './api-client';

export default function Preprocess() {
  const [requestBody, setRequestBody] = useState(JSON.stringify({
    network_identifier: { blockchain: "ethereum", network: "goerli" },
    operations: []
  }, null, 2));
  const [response, setResponse] = useState(null);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setResponse(null);
    try {
      const result = await apiClient.post('/construction/preprocess', JSON.parse(requestBody));
      setResponse(result);
    } catch (err: any) {
      setError(err.message);
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
          <Button type="submit">Preprocess</Button>
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
