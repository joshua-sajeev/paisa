"use client";

import { useEffect, useState } from "react";
import { apiFetch } from "@/lib/api";

export default function TestApiPage() {
  const [result, setResult] = useState("Testing...");

  useEffect(() => {
    apiFetch("/health")
      .then(async (response) => {
        const text = await response.text();

        if (!response.ok) {
          throw new Error(`${response.status}: ${text}`);
        }

        setResult(text);
      })
      .catch((error) => {
        setResult(`Error: ${error.message}`);
      });
  }, []);

  return (
    <main className="p-8">
      <h1 className="text-xl font-semibold">Backend connection</h1>
      <p className="mt-4">{result}</p>
    </main>
  );
}
