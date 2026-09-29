'use client';

import { useEffect, useState } from 'react';
import { getDashboardData, DashboardData } from "@/lib/dashboard";
import CalendarTimeline from "@/components/dashboard/CalendarTimeline";
import SummaryCard from "@/components/dashboard/SummmaryCard";
import AccountsList from "@/components/dashboard/AccountsList";
import JarsList from "@/components/dashboard/JarsList";
import GoalsList from "@/components/dashboard/GoalsList";
import RecentActivity from "@/components/dashboard/RecentActivity";

export default function DashboardPage() {
  const [data, setData] = useState<DashboardData | null>(null);

  useEffect(() => {
    async function loadData() {
      const result = await getDashboardData();
      setData(result);
    }
    loadData();
  }, []);

  return (
    <main className="min-h-full p-3 max-w-5xl mx-auto flex flex-col gap-3">
      <div className="flex flex-col gap-3 w-full">
        <CalendarTimeline />
        <SummaryCard summary={data?.summary ?? null} />
        <AccountsList accounts={data?.accounts ?? []} />
        <JarsList jars={data?.jars ?? []} />
        <GoalsList goals={data?.goals ?? []} />
        <RecentActivity
          transactions={data?.recent_transactions ?? []}
          onEditTransaction={(transaction) => {
            console.log('Edit:', transaction);
          }}
          onDeleteTransaction={(transaction) => {
            console.log('Delete:', transaction);
          }}
        />
      </div>
    </main>
  );
}
