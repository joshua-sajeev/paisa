"use client";

import TransactionsList from "@/components/transactions/TransactionsList";

export default function TransactionsPage() {
  return (
    <main className="min-h-screen bg-[#f8f9ff] pb-28">
      <header className="sticky top-0 z-30 bg-[#f8f9ff]/90 px-4 py-4 backdrop-blur-xl">
        <h1 className="text-2xl font-bold tracking-tight text-[#0b1c30]">
          Transactions
        </h1>

        <p className="mt-1 text-sm text-[#737686]">
          Your complete money activity
        </p>
      </header>

      <div className="px-4">
        <TransactionsList />
      </div>
    </main>
  );
}
