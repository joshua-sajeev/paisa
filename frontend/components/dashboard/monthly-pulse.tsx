import Image from "next/image";
import type { DashboardSummary } from "@/lib/types/dashboard";
import { formatMoney } from "@/lib/format";

type MonthlyPulseProps = {
  data: DashboardSummary;

};

export function MonthlyPulse({ data }: MonthlyPulseProps) {
  const month = new Intl.DateTimeFormat("en-IN", {
    month: "long",
    year: "numeric",
  }).format(new Date());

  return (
    <section className="mt-5">
      <div className="mb-2 flex items-center justify-between px-1">
        <div className="flex items-center gap-1.5">
          <span className="text-xs font-bold uppercase tracking-tight text-slate-900">
            Monthly Pulse
          </span>
        </div>

        <span className="rounded-full bg-indigo-50 px-2 py-0.5 text-[11px] font-bold text-indigo-600">
          {month}
        </span>
      </div>

      <div className="flex gap-2.5 overflow-x-auto pb-2 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
        {/* Income */}
        <div className="relative flex min-w-[145px] flex-1 flex-col overflow-hidden rounded-2xl border border-slate-100 bg-white p-3 shadow-sm">
          <div className="absolute -right-2 -top-2 h-12 w-12 rounded-bl-2xl bg-emerald-50" />

          <div className="relative z-10 mb-2 flex items-center justify-between">
            <span className="flex h-6 w-6 items-center justify-center rounded-lg bg-emerald-100 text-xs font-bold text-emerald-700">
              ↓
            </span>

            <span className="rounded bg-emerald-50 px-1 text-[10px] font-bold text-emerald-600">
              In
            </span>
          </div>

          <span className="relative z-10 text-sm font-extrabold tracking-tight text-slate-900">
            {formatMoney(data.monthly_income)}
          </span>

          <span className="relative z-10 mt-0.5 text-[11px] font-medium text-slate-400">
            Income
          </span>
        </div>

        {/* Expense */}
        <div className="relative flex min-w-[145px] flex-1 flex-col overflow-hidden rounded-2xl border border-slate-100 bg-white p-3 shadow-sm">
          <div className="absolute -right-2 -top-2 h-12 w-12 rounded-bl-2xl bg-rose-50" />

          <div className="relative z-10 mb-2 flex items-center justify-between">
            <span className="flex h-6 w-6 items-center justify-center rounded-lg bg-rose-100 text-xs font-bold text-rose-600">
              ↑
            </span>

            <span className="rounded bg-rose-50 px-1 text-[10px] font-bold text-rose-600">
              Out
            </span>
          </div>

          <span className="relative z-10 text-sm font-extrabold tracking-tight text-slate-900">
            {formatMoney(data.monthly_expense)}
          </span>

          <span className="relative z-10 mt-0.5 text-[11px] font-medium text-slate-400">
            Spent
          </span>
        </div>

        {/* Savings */}
        <div className="relative flex min-w-[145px] flex-1 flex-col overflow-hidden rounded-2xl border border-amber-100 bg-gradient-to-b from-amber-50/70 to-orange-50/50 p-3 shadow-sm">
          <div className="relative z-10 mb-2 flex items-center justify-between">
            <span className="flex h-6 w-6 items-center justify-center rounded-lg bg-amber-200">
              <Image
                src="/icons/savings.svg"
                alt=""
                width={15}
                height={15}
              />
            </span>

            <span className="rounded-full bg-amber-200/80 px-1.5 py-0.5 text-[10px] font-extrabold text-amber-800">
              {data.monthly_savings_change.toFixed(0)}%
            </span>
          </div>

          <span className="relative z-10 text-sm font-extrabold tracking-tight text-amber-950">
            {formatMoney(data.monthly_savings)}
          </span>

          <span className="relative z-10 mt-0.5 text-[11px] font-semibold text-amber-700/80">
            Saved
          </span>
        </div>
      </div>
    </section>
  );
}
