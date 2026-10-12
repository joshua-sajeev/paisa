'use client';

import { DashboardTransaction } from '@/lib/dashboard';
import { ChevronRight } from '@material-symbols-svg/react/w400';
import Link from 'next/link';
import TransactionRow from './TransactionRow';

interface RecentActivityProps {
  transactions: DashboardTransaction[];
  onEditTransaction?: (
    transaction: DashboardTransaction,
  ) => void;
  onDeleteTransaction?: (
    transaction: DashboardTransaction,
  ) => void;
}

export default function RecentActivity({
  transactions,
  onEditTransaction,
  onDeleteTransaction,
}: RecentActivityProps) {
  return (
    <section className="flex flex-col gap-2.5">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h2 className="text-[15px] font-bold text-slate-900 tracking-tight">
          Recent Activity
        </h2>

        <Link
          href="/transactions"
          className="text-xs font-semibold text-[#067FF3] flex items-center gap-0.5 hover:underline"
        >
          View All

          <ChevronRight className="w-3.5 h-3.5" />
        </Link>
      </div>

      {/* Transactions */}
      <div className="bg-white rounded-2xl border border-gray-200/80 p-2 overflow-hidden shadow-sm transition-all hover:shadow-md">
        {transactions.length === 0 ? (
          <div className="flex items-center justify-center py-8">
            <span className="text-sm text-slate-400">
              No recent transactions
            </span>
          </div>
        ) : (
          <div className="flex flex-col divide-y divide-slate-100">
            {transactions.map((transaction) => (
              <TransactionRow
                key={transaction.id}
                transaction={transaction}
                onEdit={onEditTransaction}
                onDelete={onDeleteTransaction}
              />
            ))}
          </div>
        )}
      </div>
    </section>
  );
}
