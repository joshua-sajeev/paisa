"use client";

import Link from "next/link";
import type { DashboardAccount } from "@/lib/types/dashboard";

type AccountsProps = {
  accounts: DashboardAccount[];
};

const accountStyles = [
  {
    card: "from-white to-indigo-50/70 border-indigo-100/80",
    icon: "bg-indigo-600 shadow-indigo-500/20",
    badge: "text-indigo-700 bg-indigo-100/80",
  },
  {
    card: "from-white to-emerald-50/70 border-emerald-100/80",
    icon: "bg-emerald-600 shadow-emerald-500/20",
    badge: "text-emerald-800 bg-emerald-100",
  },
  {
    card: "from-white to-amber-50/70 border-amber-100/80",
    icon: "bg-amber-500 shadow-amber-500/20",
    badge: "text-amber-800 bg-amber-100",
  },
];

function formatMoney(amount: number) {
  return new Intl.NumberFormat("en-IN", {
    style: "currency",
    currency: "INR",
    minimumFractionDigits: 2,
  }).format(amount / 100);
}

function BankIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      className="h-5 w-5"
      aria-hidden="true"
    >
      <path d="M3 10h18" />
      <path d="M5 10v8" />
      <path d="M9 10v8" />
      <path d="M15 10v8" />
      <path d="M19 10v8" />
      <path d="M2 18h20" />
      <path d="M12 3 2 8h20L12 3Z" />
    </svg>
  );
}

export function Accounts({ accounts }: AccountsProps) {
  return (
    <section className="flex flex-col space-y-2 px-4">
      <div className="flex items-center justify-between px-1">
        <h2 className="text-sm font-extrabold tracking-tight text-slate-900">
          Active Accounts
        </h2>

        <Link
          href="/accounts"
          className="flex items-center gap-0.5 text-xs font-bold text-indigo-600 hover:text-indigo-700"
        >
          View All ({accounts.length})

          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2.5"
            strokeLinecap="round"
            strokeLinejoin="round"
            className="h-3.5 w-3.5"
            aria-hidden="true"
          >
            <path d="m9 18 6-6-6-6" />
          </svg>
        </Link>
      </div>

      <div className="-mx-4 flex space-x-3 overflow-x-auto px-4 pb-1 scrollbar-none">
        {accounts.map((account, index) => {
          const style = accountStyles[index % accountStyles.length];

          return (
            <div
              key={account.id}
              className={`flex w-64 flex-shrink-0 flex-col justify-between rounded-3xl border bg-gradient-to-br p-4 shadow-sm transition-all hover:shadow-md ${style.card}`}
            >
              <div className="mb-3 flex items-center justify-between">
                <div className="flex items-center gap-2.5">
                  <div
                    className={`flex h-10 w-10 items-center justify-center rounded-2xl text-white shadow-md ${style.icon}`}
                  >
                    <BankIcon />
                  </div>

                  <div className="flex flex-col">
                    <span className="text-xs font-extrabold text-slate-900">
                      {account.name}
                    </span>
                  </div>
                </div>

                <span
                  className={`rounded-full px-2 py-0.5 text-[10px] font-black ${style.badge}`}
                >
                  Active
                </span>
              </div>

              <div>
                <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">
                  Available Balance
                </span>

                <p className="mt-0.5 text-xl font-extrabold tracking-tight text-slate-900">
                  {formatMoney(account.balance)}
                </p>
              </div>
            </div>
          );
        })}
      </div>
    </section>
  );
}
