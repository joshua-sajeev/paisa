'use client';

import { DashboardSummary } from '@/lib/dashboard';
import { formatMoney } from '@/lib/utils';
import { usePrivacy } from '@/context/PrivacyContext';

interface SummaryCardProps {
  summary: DashboardSummary | null;
}

export default function SummaryCard({ summary }: SummaryCardProps) {
  const { isPrivate } = usePrivacy();

  return (
    <div className="w-full bg-white p-6 rounded-[22px] border border-gray-200/80 shadow-sm transition-all hover:shadow-md">
      <p className="text-gray-400 text-xs font-semibold tracking-wider">TOTAL NET WORTH</p>
      <h2 className="text-2xl font-bold mt-1 text-[#1c1b1f]">
        {isPrivate ? '••••••' : `₹${formatMoney(summary?.total_balance ?? 0)}`}
      </h2>
    </div>
  );
}
