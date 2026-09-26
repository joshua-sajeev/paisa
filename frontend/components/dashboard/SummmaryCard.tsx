'use client';

import { DashboardSummary } from '@/lib/dashboard';
import { formatMoney } from '@/lib/utils';
import { usePrivacy } from '@/context/PrivacyContext';
import {
  ArrowUpward,
  ArrowDownward,
  Savings,
} from '@material-symbols-svg/react/w400';

interface SummaryCardProps {
  summary: DashboardSummary | null;
}

export default function SummaryCard({ summary }: SummaryCardProps) {
  const { isPrivate } = usePrivacy();

  return (
    <section className="relative overflow-hidden bg-white rounded-xl p-5 border border-gray-200/80 shadow-sm transition-all hover:shadow-md flex flex-col gap-4">
      {/* Decorative background blurs */}
      <div className="absolute -top-12 -right-12 w-36 h-36 rounded-full bg-amber-100 opacity-40 blur-2xl pointer-events-none" />
      <div className="absolute -bottom-10 -left-10 w-32 h-32 rounded-full bg-orange-100 opacity-30 blur-2xl pointer-events-none" />

      {/* Top Net Worth Section */}
      <div className="relative z-10 flex flex-col gap-1">
        <span className="text-[11px] font-semibold text-gray-500 uppercase tracking-wider">
          Total Net Worth
        </span>

        <span className="text-[24px] leading-tight text-neutral-900 tracking-tight font-extrabold">
          {isPrivate
            ? '••••••••'
            : `₹${formatMoney(summary?.total_balance ?? 0)}`}
        </span>
      </div>

      {/* Breakdown Section */}
      <div className="relative z-10 flex flex-col gap-2 pt-1">
        <div className="grid grid-cols-2 gap-2">
          {/* Monthly Income */}
          <div className="flex flex-col gap-0.5 p-3 rounded-lg bg-[rgb(236,253,245)]">
            <div className="flex items-center">
              <div className="w-6 h-6 rounded-md flex items-center justify-center flex-shrink-0 bg-[rgba(16,185,129,0.2)] text-[rgb(5,150,105)]">
                <ArrowUpward width={14} height={14} />
              </div>
            </div>

            <div className="flex flex-col min-w-0 mt-1">
              <span className="text-[10px] font-medium leading-tight text-[rgb(5,150,105)]">
                Monthly Income
              </span>

              <span className="text-[14px] font-bold leading-tight text-[rgb(5,150,105)]">
                {isPrivate
                  ? '••••••'
                  : `₹${formatMoney(summary?.monthly_income ?? 0)}`}
              </span>
            </div>
          </div>

          {/* Monthly Spend */}
          <div className="flex flex-col gap-0.5 p-3 rounded-lg bg-[rgb(254,242,242)]">
            <div className="flex items-center">
              <div className="w-6 h-6 rounded-md flex items-center justify-center flex-shrink-0 bg-[rgba(239,68,68,0.18)] text-[rgb(220,38,38)]">
                <ArrowDownward width={14} height={14} />
              </div>
            </div>

            <div className="flex flex-col min-w-0 mt-1">
              <span className="text-[10px] font-medium leading-tight text-[rgb(220,38,38)]">
                Monthly Spend
              </span>

              <span className="text-[14px] font-bold leading-tight text-[rgb(220,38,38)]">
                {isPrivate
                  ? '••••••'
                  : `₹${formatMoney(summary?.monthly_expense ?? 0)}`}
              </span>
            </div>
          </div>
        </div>

        {/* Monthly Savings Bar */}
        <div className="flex items-center justify-between px-3 py-2 rounded-lg bg-[rgb(238,242,255)]">
          <div className="flex items-center gap-2">
            <div className="w-6 h-6 rounded-md flex items-center justify-center flex-shrink-0 bg-[rgba(99,102,241,0.2)] text-[rgb(55,48,163)]">
              <Savings width={14} height={14} />
            </div>

            <div className="flex items-center gap-1.5">
              <span className="text-[11px] font-medium text-[rgb(55,48,163)]">
                Monthly Savings:
              </span>

              <span className="text-[13px] font-bold text-[rgb(55,48,163)]">
                {isPrivate
                  ? '••••••'
                  : `₹${formatMoney(summary?.monthly_savings ?? 0)}`}
              </span>
            </div>
          </div>

          <span
            className="text-[10px] px-1.5 py-0.5 rounded-full font-bold"
            style={{
              backgroundColor: 'rgba(99, 102, 241, 0.15)',
              color: 'rgb(55, 48, 163)',
              border: '1px solid rgba(99, 102, 241, 0.3)',
            }}
          >
            {summary?.monthly_savings_rate?.toFixed(1) ?? '0.0'}%
          </span>
        </div>
      </div>
    </section>
  );
}
