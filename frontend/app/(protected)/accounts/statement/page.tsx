"use client";

import { useSearchParams } from "next/navigation";

import AccountStatement from "@/components/accounts/AccountStatement";

export default function AccountStatementPage() {
  const searchParams = useSearchParams();
  const accountId = searchParams.get("accountId");

  if (!accountId) {
    return (
      <main className="flex min-h-screen items-center justify-center px-4">
        <p className="text-sm font-medium text-error">
          Account not found.
        </p>
      </main>
    );
  }

  return <AccountStatement accountId={accountId} />;
}
