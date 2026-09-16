"use client";

import Link from "next/link";
import type { DashboardTransaction } from "@/lib/types/dashboard";

type RecentTransactionsProps = {
  transactions: DashboardTransaction[];
};

function formatMoney(paise: number): string {
  return new Intl.NumberFormat("en-IN", {
    style: "currency",
    currency: "INR",
    maximumFractionDigits: 2,
  }).format(paise / 100);
}

function formatDate(date: string): string {
  const parsed = new Date(`${date}T00:00:00`);

  return new Intl.DateTimeFormat("en-IN", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  }).format(parsed);
}

function getTransactionStyle(
  transaction: DashboardTransaction,
) {
  switch (transaction.type) {
    case "income":
      return {
        icon: (
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
            <path d="M12 2v20" />
            <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H7" />
          </svg>
        ),
        background: "bg-emerald-100",
        text: "text-emerald-700",
        amount: "text-emerald-600",
        sign: "+",
      };

    case "transfer":
      return {
        icon: (
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
            <path d="m17 3 4 4-4 4" />
            <path d="M3 7h18" />
            <path d="m7 21-4-4 4-4" />
            <path d="M21 17H3" />
          </svg>
        ),
        background: "bg-indigo-100",
        text: "text-indigo-700",
        amount: "text-indigo-600",
        sign: "",
      };

    default:
      return {
        icon: (
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
            <path d="M6 2h9l5 5v15H6z" />
            <path d="M14 2v6h6" />
            <path d="M9 13h6" />
            <path d="M9 17h4" />
          </svg>
        ),
        background: "bg-rose-100",
        text: "text-rose-600",
        amount: "text-rose-600",
        sign: "-",
      };
  }
}

export function RecentTransactions({
  transactions,
}: RecentTransactionsProps) {
  return (
    <section className="flex flex-col space-y-2 pb-6">
      <div className="flex items-center justify-between px-1">
        <div className="flex items-center gap-1.5">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            className="h-4 w-4 text-indigo-600"
            aria-hidden="true"
          >
            <path d="M6 2h9l5 5v15H6z" />
            <path d="M14 2v6h6" />
            <path d="M9 13h6" />
            <path d="M9 17h4" />
          </svg>

          <h2 className="text-sm font-extrabold tracking-tight text-slate-900">
            Recent Activity
          </h2>
        </div>

        <Link
          href="/transactions"
          className="text-xs font-bold text-indigo-600 hover:underline"
        >
          History
        </Link>
      </div>

      <div className="space-y-1 rounded-3xl border border-slate-100 bg-white p-2 shadow-sm">
        {transactions.length === 0 ? (
          <div className="px-3 py-8 text-center">
            <p className="text-sm font-semibold text-slate-500">
              No recent transactions
            </p>
          </div>
        ) : (
          transactions.map((transaction) => {
            const style = getTransactionStyle(transaction);

            return (
              <div
                key={transaction.id}
                className="group flex items-center justify-between rounded-2xl p-2.5 transition-colors hover:bg-slate-50"
              >
                <div className="flex min-w-0 items-center gap-3">
                  <div
                    className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl font-bold transition-transform group-hover:scale-105 ${style.background} ${style.text}`}
                  >
                    {style.icon}
                  </div>

                  <div className="flex min-w-0 flex-col">
                    <span className="truncate text-xs font-bold leading-snug text-slate-900">
                      {transaction.name}
                    </span>

                    <div className="mt-0.5 flex items-center gap-1.5 text-[10px] font-medium text-slate-400">
                      <span className="truncate">
                        {transaction.account}
                      </span>

                      <span className="h-1 w-1 shrink-0 rounded-full bg-slate-300" />

                      <span className="shrink-0">
                        {formatDate(transaction.date)}
                      </span>
                    </div>
                  </div>
                </div>

                <span
                  className={`ml-3 shrink-0 text-sm font-extrabold tracking-tight ${style.amount}`}
                >
                  {style.sign}
                  {formatMoney(transaction.amount)}
                </span>
              </div>
            );
          })
        )}
      </div>
    </section>
  );
}
