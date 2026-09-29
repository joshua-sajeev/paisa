'use client';

import { useState } from 'react';
import {
  ReceiptLong,
  Movie,
  TrendingUp,
  VolunteerActivism,
  Tune,
  Savings,
  type MaterialSymbolsComponent,
} from '@material-symbols-svg/react/w400';

import { formatMoney } from '@/lib/utils';
import { usePrivacy } from '@/context/PrivacyContext';

interface Jar {
  id: string;
  name: string;
  allocation_type: 'fixed' | 'percentage' | 'remainder';
  allocation_value: number;
  allocated: number;
  used: number;
  available: number;
  used_percentage: number;
  available_percentage: number;
}

interface JarsListProps {
  jars: Jar[];
}

const JAR_COLORS = [
  '#F43F5E',
  '#6366F1',
  '#07B682',
  '#F6A723',
];

const getJarIcon = (name: string): MaterialSymbolsComponent => {
  const key = name.toLowerCase();

  if (key.includes('necessit')) return ReceiptLong;
  if (key.includes('leisure')) return Movie;
  if (key.includes('invest')) return TrendingUp;
  if (key.includes('giv')) return VolunteerActivism;

  return Savings;
};

function getAllocationLabel(jar: Jar) {
  switch (jar.allocation_type) {
    case 'percentage':
      return `${jar.allocation_value}% of income`;

    case 'fixed':
      return `₹${formatMoney(jar.allocation_value)} fixed`;

    case 'remainder':
      return 'Remainder';
  }
}

function getStatus(jar: Jar) {
  if (jar.available <= 0) {
    return {
      label: 'Fully Spent',
      className:
        'rounded-full border border-rose-200 bg-rose-50 px-1.5 py-0.5 text-[10px] font-semibold text-rose-700',
    };
  }

  return {
    label: `${Math.round(jar.available_percentage)}% Left`,
    className:
      'rounded-full border border-emerald-200 bg-emerald-50 px-1.5 py-0.5 text-[10px] font-semibold text-emerald-700',
  };
}

export default function JarsList({ jars }: JarsListProps) {
  const { isPrivate } = usePrivacy();
  const [revealedIds, setRevealedIds] = useState<Set<string>>(new Set());

  if (!jars || jars.length === 0) return null;

  const toggleReveal = (id: string) => {
    if (!isPrivate) return;

    setRevealedIds((prev) => {
      const next = new Set(prev);

      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }

      return next;
    });
  };

  const formatAmount = (amount: number, jarId: string) => {
    const isRevealed = revealedIds.has(jarId);

    if (isPrivate && !isRevealed) {
      return (
        <span className="inline-flex items-center rounded-md border border-slate-200/70 bg-slate-100/70 px-1.5 py-0.5 text-slate-400 backdrop-blur-sm">
          ₹••••
        </span>
      );
    }

    return <>₹{formatMoney(amount)}</>;
  };

  return (
    <section className="flex flex-col gap-2.5">
      <div className="flex items-center justify-between">
        <div className="flex flex-col">
          <h2 className="text-[15px] font-bold tracking-tight text-on-surface">
            Smart Jars
          </h2>

          <span className="text-[11px] font-medium text-on-surface-variant">
            Monthly budget allocations
          </span>
        </div>

        <button
          type="button"
          aria-label="Jar budget filter and settings"
          className="flex items-center justify-center rounded-lg border border-slate-200 bg-slate-50 p-1.5 text-slate-600 transition-colors hover:bg-slate-100"
        >
          <Tune size={14} color="currentColor" />
        </button>
      </div>

      <div className="flex flex-col gap-2">
        {jars.map((jar, index) => {
          const Icon = getJarIcon(jar.name);
          const color = JAR_COLORS[index % JAR_COLORS.length];
          const status = getStatus(jar);

          return (
            <button
              key={jar.id}
              type="button"
              onClick={() => toggleReveal(jar.id)}
              className="flex flex-col gap-2 rounded-xl bg-white p-3 text-left shadow-sm transition-transform active:scale-[0.99]"
              aria-label={
                isPrivate
                  ? `Toggle amount visibility for ${jar.name}`
                  : `${jar.name} jar`
              }
            >
              {/* Header */}
              <div className="flex items-center justify-between">
                <div className="flex min-w-0 items-center gap-2">
                  <div
                    className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg"
                    style={{
                      backgroundColor: `${color}18`,
                    }}
                  >
                    <Icon size={14} color={color} />
                  </div>

                  <span className="truncate text-sm font-bold text-on-surface">
                    {jar.name}
                  </span>
                </div>

                <span className="ml-2 shrink-0 rounded-full border border-slate-200 bg-slate-100 px-2 py-0.5 text-[10px] font-medium text-slate-700">
                  {getAllocationLabel(jar)}
                </span>
              </div>

              {/* Progress */}
              <div className="flex flex-col gap-1">
                <span className="flex items-center gap-1 text-[11px] font-medium text-slate-500">
                  {formatAmount(jar.used, jar.id)}
                  <span>spent of</span>
                  {formatAmount(jar.allocated, jar.id)}
                </span>

                <div className="h-1.5 w-full overflow-hidden rounded-full bg-slate-100">
                  <div
                    className="h-full rounded-full transition-all"
                    style={{
                      width: `${Math.min(jar.used_percentage, 100)}%`,
                      backgroundColor: color,
                    }}
                  />
                </div>
              </div>

              {/* Bottom */}
              <div className="flex items-center justify-between">
                <span className="flex items-center gap-1 text-xs font-bold text-on-surface">
                  {formatAmount(jar.available, jar.id)}
                  <span>Available</span>
                </span>

                <span className={status.className}>{status.label}</span>
              </div>
            </button>
          );
        })}
      </div>
    </section>
  );
}
