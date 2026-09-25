'use client';

import { useEffect, useState } from 'react';
import CalendarTimeline from "@/components/dashboard/CalendarTimeline";
import SummaryCard from "@/components/dashboard/SummmaryCard";
import { getDashboardData, DashboardData } from "@/lib/dashboard";

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
      </div>
    </main>
  );
}
