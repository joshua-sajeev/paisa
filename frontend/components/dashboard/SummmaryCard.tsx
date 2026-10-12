'use client';

import { useEffect, useRef, useState } from 'react';
import Link from 'next/link';
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
  const [isRevealed, setIsRevealed] = useState(false);
  const pressTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const longPressTriggered = useRef(false);

  const toggleReveal = () => {
    if (!isPrivate) return;
    setIsRevealed((prev) => !prev);
  };

  const clearPressTimer = () => {
    if (pressTimer.current) clearTimeout(pressTimer.current);
    pressTimer.current = null;
  };

  useEffect(() => clearPressTimer, []);

  const formatAmount = (
    amount: number,
    size: 'large' | 'small' = 'small',
  ) => {
    if (isPrivate && !isRevealed) {
      return (
        <span
          className={size === 'large' ? 'font-extrabold tracking-widest text-slate-900' : 'font-bold tracking-widest text-slate-900'}
        >
          ₹••••
        </span>
      );
    }

    return <>₹{formatMoney(amount)}</>;
  };

  return (
    <div
      onPointerDown={(event) => {
        if (event.button !== 0) return;
        clearPressTimer();
        longPressTriggered.current = false;
        pressTimer.current = setTimeout(() => {
          if (isPrivate) {
            longPressTriggered.current = true;
            toggleReveal();
          }
        }, 500);
      }}
      onPointerUp={clearPressTimer}
      onPointerLeave={clearPressTimer}
      onPointerCancel={() => {
        clearPressTimer();
        longPressTriggered.current = false;
      }}
      onContextMenu={(event) => event.preventDefault()}
      onClickCapture={(event) => {
        if (longPressTriggered.current) {
          event.preventDefault();
          event.stopPropagation();
          longPressTriggered.current = false;
        }
      }}
      className="relative flex w-full touch-manipulation select-none flex-col gap-4 overflow-hidden rounded-xl border border-gray-200/80 bg-white p-5 text-left shadow-sm transition-all hover:shadow-md"
      aria-label={isPrivate ? 'Press and hold to reveal summary amounts' : undefined}
      aria-description={isPrivate ? "Press and hold for half a second to toggle amount visibility." : undefined}
    >
      {/* Decorative background blurs */}
      <div className="pointer-events-none absolute -right-12 -top-12 h-36 w-36 rounded-full bg-amber-100 opacity-40 blur-2xl" />
      <div className="pointer-events-none absolute -bottom-10 -left-10 h-32 w-32 rounded-full bg-orange-100 opacity-30 blur-2xl" />

      {/* Total Net Worth */}
      <div className="relative z-10 flex flex-col gap-1">
        <span className="text-[11px] font-semibold uppercase tracking-wider text-gray-500">
          Total Net Worth
        </span>

        <span className="text-[24px] font-extrabold leading-tight tracking-tight text-neutral-900">
          {formatAmount(summary?.total_balance ?? 0, 'large')}
        </span>
      </div>

      {/* Breakdown */}
      <div className="relative z-10 flex flex-col gap-2 pt-1">
        <div className="grid grid-cols-2 gap-2">
          {/* Monthly Income */}
          <Link href="/transactions?type=income" className="flex flex-col gap-0.5 rounded-lg bg-[rgb(236,253,245)] p-3 transition hover:brightness-[0.98]">
            <div className="flex items-center">
              <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-md bg-[rgba(16,185,129,0.2)] text-[rgb(5,150,105)]">
                <ArrowUpward size={14} color="currentColor" />
              </div>
            </div>

            <div className="mt-1 flex min-w-0 flex-col">
              <span className="text-[10px] font-medium leading-tight text-[rgb(5,150,105)]">
                Monthly Income
              </span>

              <span className="text-[14px] font-bold leading-tight text-[rgb(5,150,105)]">
                {formatAmount(summary?.monthly_income ?? 0)}
              </span>
            </div>
          </Link>

          {/* Monthly Spend */}
          <Link href="/transactions?type=expense" className="flex flex-col gap-0.5 rounded-lg bg-[rgb(254,242,242)] p-3 transition hover:brightness-[0.98]">
            <div className="flex items-center">
              <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-md bg-[rgba(239,68,68,0.18)] text-[rgb(220,38,38)]">
                <ArrowDownward size={14} color="currentColor" />
              </div>
            </div>

            <div className="mt-1 flex min-w-0 flex-col">
              <span className="text-[10px] font-medium leading-tight text-[rgb(220,38,38)]">
                Monthly Spend
              </span>

              <span className="text-[14px] font-bold leading-tight text-[rgb(220,38,38)]">
                {formatAmount(summary?.monthly_expense ?? 0)}
              </span>
            </div>
          </Link>
        </div>

        {/* Monthly Savings */}
        <Link href="/transactions" className="flex items-center justify-between rounded-lg bg-[rgb(238,242,255)] px-3 py-2 transition hover:brightness-[0.98]">
          <div className="flex items-center gap-2">
            <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-md bg-[rgba(99,102,241,0.2)] text-[rgb(55,48,163)]">
              <Savings size={14} color="currentColor" />
            </div>

            <div className="flex items-center gap-1.5">
              <span className="text-[11px] font-medium text-[rgb(55,48,163)]">
                Monthly Savings:
              </span>

              <span className="text-[13px] font-bold text-[rgb(55,48,163)]">
                {formatAmount(summary?.monthly_savings ?? 0)}
              </span>
            </div>
          </div>

          {/* Percentage stays visible */}
          <span
            className="rounded-full px-1.5 py-0.5 text-[10px] font-bold"
            style={{
              backgroundColor: 'rgba(99, 102, 241, 0.15)',
              color: 'rgb(55, 48, 163)',
              border: '1px solid rgba(99, 102, 241, 0.3)',
            }}
          >
            {summary?.monthly_savings_rate?.toFixed(1) ?? '0.0'}%
          </span>
        </Link>
      </div>
    </div>
  );
}
