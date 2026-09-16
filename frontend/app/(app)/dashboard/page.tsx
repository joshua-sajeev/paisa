"use client";

import Link from "next/link";
import Image from "next/image";
import { useEffect, useState } from "react";
import { Summary } from "@/components/dashboard/summary";
import type { DashboardData } from "@/lib/types/dashboard";
import { Accounts } from "@/components/dashboard/accounts";
import { Jars } from "@/components/dashboard/jars";
import { Goals } from "@/components/dashboard/goals";
import { RecentTransactions } from "@/components/dashboard/recent-transactions";

export default function DashboardPage() {
  const [data, setData] = useState<DashboardData | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadDashboard() {
      try {
        const response = await fetch("/api/v1/dashboard", {
          credentials: "include",
          cache: "no-store",
        });

        if (!response.ok) {
          throw new Error(`Dashboard request failed: ${response.status}`);
        }

        const dashboard: DashboardData = await response.json();
        setData(dashboard);
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "Failed to load dashboard",
        );
      }
    }

    loadDashboard();
  }, []);

  if (error) {
    return (
      <main>
        <p>{error}</p>
      </main>
    );
  }

  if (!data) {
    return (
      <main>
        <p>Loading...</p>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-[#f6f8fd]">
      <header className="flex h-12 items-center px-5 pt-8">
        <Link
          href="/dashboard"
          aria-label="Go to dashboard"
          className="flex items-center"
        >
          <Image
            src="/icons/paisa-logo.svg"
            alt="PAISA"
            width={36}
            height={36}
            priority
          />
          <span className="ml-2 text-lg font-extrabold tracking-tight">
            PAISA
          </span>
        </Link>
      </header>

      <div className="-mt-1">
        <Summary data={data.summary} />
      </div>

      <div className="mt-5">
        <Accounts accounts={data.accounts} />
      </div>

      <div className="mt-6">
        <Jars jars={data.jars} />
      </div>

      <div className="mt-6">
        <Goals goals={data.goals} />
      </div>

      <div className="mt-6">
        <RecentTransactions
          transactions={data.recent_transactions}
        />
      </div>
    </main>

  );
}
