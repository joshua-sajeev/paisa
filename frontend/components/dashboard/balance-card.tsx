"use client";

import { useState } from "react";
import type { DashboardSummary } from "@/lib/types/dashboard";
import { formatMoney } from "@/lib/format";

type BalanceCardProps = {
  data: DashboardSummary;
};

export function BalanceCard({ data }: BalanceCardProps) {
  const [visible, setVisible] = useState(false);

  return (
    <section className="relative overflow-hidden rounded-3xl bg-gradient-to-br from-[#1E1B4B] via-[#312E81] to-[#4338CA] p-5 text-white shadow-xl shadow-indigo-500/5">
      {/* Decorative blobs */}
      <div className="pointer-events-none absolute -right-10 -top-10 h-36 w-36 rounded-full bg-[#FF5C67]/30 blur-2xl" />
      <div className="pointer-events-none absolute -bottom-8 -left-6 h-32 w-32 rounded-full bg-[#FFB800]/20 blur-xl" />

      <div className="relative z-10">
        {/* Top tag bar */}
        <div className="flex items-center justify-between">
          <div className="inline-flex items-center gap-1.5 rounded-full border border-white/15 bg-white/10 px-3 py-1 backdrop-blur-md">
            <span className="text-[11px] font-bold uppercase tracking-wider text-indigo-100">
              Total Wealth
            </span>

            <button
              type="button"
              aria-label={visible ? "Hide balance" : "Show balance"}
              onClick={() => setVisible((current) => !current)}
              className="ml-0.5 flex h-5 w-5 items-center justify-center rounded-full text-indigo-200 transition-colors hover:text-white"
            >
              <span className="flex h-5 w-5 items-center justify-center">
                {visible ? (
                  <svg
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    aria-hidden="true"
                  >
                    <path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12Z" />
                    <circle cx="12" cy="12" r="3" />
                  </svg>
                ) : (
                  <svg
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    aria-hidden="true"
                  >
                    <path d="M3 3l18 18" />
                    <path d="M10.6 10.6a2 2 0 0 0 2.8 2.8" />
                    <path d="M9.9 4.2A10.7 10.7 0 0 1 12 4c6.5 0 10 8 10 8a17.4 17.4 0 0 1-3.2 4.5" />
                    <path d="M6.2 6.2C3.7 8 2 12 2 12s3.5 8 10 8c1.1 0 2.1-.2 3-.5" />
                  </svg>
                )}
              </span>
            </button>
          </div>

          <div className="inline-flex items-center gap-1 rounded-full border border-emerald-400/30 bg-emerald-400/20 px-2.5 py-1 text-[11px] font-bold text-emerald-300">
            <span>{data.monthly_savings_change.toFixed(1)}% saved</span>
          </div>
        </div>

        {/* Balance */}
        <div className="mt-4">
          <h1 className="text-4xl font-extrabold leading-none tracking-tight">
            {visible ? formatMoney(data.total_balance) : "••••••"}
          </h1>
        </div>
      </div>
    </section>
  );
}
