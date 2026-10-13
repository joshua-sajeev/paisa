'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { getDashboardData, DashboardData, DashboardTransaction } from "@/lib/dashboard";
import DeleteTransactionDialog from "@/components/transactions/DeleteTransactionDialog";
import SummaryCard from "@/components/dashboard/SummmaryCard";
import AccountsList from "@/components/dashboard/AccountsList";
import JarsList from "@/components/dashboard/JarsList";
import GoalsList from "@/components/dashboard/GoalsList";
import RecentActivity from "@/components/dashboard/RecentActivity";

export default function DashboardPage() {
  const router = useRouter();
  const [data, setData] = useState<DashboardData | null>(null);
  const [pendingDelete, setPendingDelete] = useState<DashboardTransaction | null>(null);

  useEffect(() => {
    async function loadData() {
      const result = await getDashboardData();
      setData(result);
    }
    loadData();
  }, []);

  async function refreshData() {
    setData(await getDashboardData());
  }

  return (
    <main className="min-h-full p-3 max-w-5xl mx-auto flex flex-col gap-3">
      <div className="flex flex-col gap-3 w-full">
        <SummaryCard summary={data?.summary ?? null} />
        <AccountsList accounts={data?.accounts ?? []} />
        <JarsList jars={data?.jars ?? []} />
        <GoalsList goals={data?.goals ?? []} />
        <RecentActivity
          transactions={data?.recent_transactions ?? []}
          onEditTransaction={(transaction) => {
            router.push(`/transactions/edit?id=${encodeURIComponent(transaction.id)}`);
          }}
          onDeleteTransaction={(transaction) => setPendingDelete(transaction)}
        />
      </div>

      <DeleteTransactionDialog
        transaction={pendingDelete}
        onClose={() => setPendingDelete(null)}
        onDeleted={refreshData}
      />
    </main>
  );
}
