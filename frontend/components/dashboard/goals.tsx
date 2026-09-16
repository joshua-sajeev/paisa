"use client";

import type { DashboardGoal } from "@/lib/types/dashboard";

type GoalsProps = {
  goals: DashboardGoal[];
};

function formatMoney(paise: number): string {
  return new Intl.NumberFormat("en-IN", {
    style: "currency",
    currency: "INR",
    maximumFractionDigits: 0,
  }).format(paise / 100);
}

function formatPercentage(value: number): string {
  return `${Math.round(value)}%`;
}

export function Goals({ goals }: GoalsProps) {
  return (
    <section className="flex flex-col space-y-2.5 px-4">
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
            className="h-4 w-4 text-purple-600"
            aria-hidden="true"
          >
            <circle cx="12" cy="12" r="9" />
            <circle cx="12" cy="12" r="5" />
            <circle cx="12" cy="12" r="1.5" />
          </svg>

          <h2 className="text-sm font-extrabold tracking-tight text-slate-900">
            Financial Goals
          </h2>
        </div>

        <span className="rounded-full bg-purple-100 px-2 py-0.5 text-[10px] font-black uppercase tracking-wider text-purple-700">
          {goals.length} Active
        </span>
      </div>

      <div className="grid grid-cols-1 gap-2.5">
        {goals.map((goal, index) => {
          const percentage = Math.min(
            100,
            Math.max(0, goal.progress),
          );

          const remaining = goal.remaining;

          const isGreen = index % 2 === 1;

          return (
            <div
              key={goal.id}
              className={`relative flex flex-col overflow-hidden rounded-3xl border bg-white p-4 shadow-sm ${isGreen ? "border-emerald-100" : "border-indigo-100"
                }`}
            >
              {/* Header */}
              <div className="mb-2 flex items-center justify-between">
                <div className="flex items-center gap-2.5">
                  <div
                    className={`flex h-9 w-9 items-center justify-center rounded-2xl text-white shadow-sm ${isGreen
                        ? "bg-gradient-to-tr from-emerald-500 to-teal-500"
                        : "bg-gradient-to-tr from-indigo-500 to-purple-500"
                      }`}
                  >
                    {isGreen ? (
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
                        <path d="M12 3v18" />
                        <path d="M17 8a5 5 0 0 0-5-5" />
                        <path d="M7 16a5 5 0 0 0 5 5" />
                        <path d="M17 8h-5a5 5 0 0 0-5 5" />
                      </svg>
                    ) : (
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
                        <path d="M3 21h18" />
                        <path d="M5 21V7l7-4 7 4v14" />
                        <path d="M9 21v-6h6v6" />
                        <path d="M9 9h.01" />
                        <path d="M15 9h.01" />
                      </svg>
                    )}
                  </div>

                  <div>
                    <span className="block text-xs font-extrabold text-slate-900">
                      {goal.name}
                    </span>

                    <span className="block text-[10px] text-slate-400">
                      Goal progress
                    </span>
                  </div>
                </div>

                <span
                  className={`rounded-full border px-2.5 py-1 text-xs font-black ${isGreen
                      ? "border-emerald-100 bg-emerald-50 text-emerald-600"
                      : "border-indigo-100 bg-indigo-50 text-indigo-600"
                    }`}
                >
                  {formatPercentage(percentage)}
                </span>
              </div>

              {/* Progress */}
              <div className="my-2 h-2.5 w-full overflow-hidden rounded-full bg-slate-100 p-0.5">
                <div
                  className={`h-full rounded-full ${isGreen
                      ? "bg-gradient-to-r from-emerald-400 to-teal-500"
                      : "bg-gradient-to-r from-indigo-500 via-purple-500 to-indigo-600"
                    }`}
                  style={{
                    width: `${percentage}%`,
                  }}
                />
              </div>

              {/* Amounts */}
              <div className="mt-1 flex items-baseline justify-between">
                <div className="flex items-baseline gap-1">
                  <span className="text-sm font-extrabold text-slate-900">
                    {formatMoney(goal.contributed)}
                  </span>

                  <span className="text-[11px] font-bold text-slate-400">
                    / {formatMoney(goal.target)}
                  </span>
                </div>

                <span
                  className={`rounded-md px-2 py-0.5 text-[11px] font-bold ${isGreen
                      ? "bg-emerald-50/70 text-emerald-700"
                      : "bg-indigo-50/70 text-indigo-600"
                    }`}
                >
                  {formatMoney(remaining)} to go
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </section>
  );
}
