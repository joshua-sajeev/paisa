"use client";

import Link from "next/link";
import type { DashboardJar } from "@/lib/types/dashboard";
import { formatMoney } from "@/lib/format";
import { formatPercentage } from "@/lib/format";
type JarsProps = {
  jars: DashboardJar[];
};

const jarStyles = [
  {
    icon: "bg-indigo-100 text-indigo-700",
    border: "hover:border-indigo-200",
    badge: "text-indigo-700 bg-indigo-50",
    progress: "bg-indigo-500",
    iconPath: (
      <>
        <path d="M7 4h10" />
        <path d="M8 4v3l-2 3v7a3 3 0 0 0 3 3h6a3 3 0 0 0 3-3v-7l-2-3V4" />
        <path d="M6 13h12" />
      </>
    ),
  },
  {
    icon: "bg-teal-100 text-teal-700",
    border: "hover:border-teal-200",
    badge: "text-teal-700 bg-teal-50",
    progress: "bg-teal-500",
    iconPath: (
      <>
        <path d="M2 16h20" />
        <path d="M4 16l2-6h12l2 6" />
        <path d="M6 10l2-4" />
        <path d="M18 10l-2-4" />
        <path d="M12 6v4" />
      </>
    ),
  },
  {
    icon: "bg-amber-100 text-amber-700",
    border: "hover:border-amber-200",
    badge: "text-amber-800 bg-amber-50",
    progress: "bg-amber-500",
    iconPath: (
      <>
        <path d="M4 12h16" />
        <path d="M5 12v5" />
        <path d="M19 12v5" />
        <path d="M3 17h18" />
        <path d="M6 17v2" />
        <path d="M18 17v2" />
        <path d="M7 12V9h10v3" />
      </>
    ),
  },
];


function JarIcon({ children }: { children: React.ReactNode }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      className="h-[18px] w-[18px]"
      aria-hidden="true"
    >
      {children}
    </svg>
  );
}

function getAllocationLabel(jar: DashboardJar) {
  switch (jar.allocation_type) {
    case "percentage":
      return `${formatPercentage(jar.allocation_value)} of income`;

    case "fixed":
      return `${formatMoney(jar.allocation_value)} /mo`;

    case "remainder":
      return "Spare cash";

    default:
      return "";
  }
}

export function Jars({ jars }: JarsProps) {
  return (
    <section className="flex flex-col space-y-2.5 px-4">
      <div className="flex items-center justify-between px-1">
        <div>
          <div className="flex items-center gap-1.5">
            <h2 className="text-sm font-extrabold tracking-tight text-slate-900">
              Smart Jars
            </h2>
          </div>
          <p className="text-[11px] font-medium text-slate-400">
            Automatic rules & rainy day stashes
          </p>
        </div>

        <Link
          href="/jars"
          aria-label="Manage jars"
          className="flex h-8 w-8 items-center justify-center rounded-full border border-slate-100 bg-white text-slate-600 shadow-sm transition-colors hover:border-indigo-100 hover:text-indigo-600"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            className="h-4 w-4"
            aria-hidden="true"
          >
            <path d="M4 6h16" />
            <path d="M4 12h16" />
            <path d="M4 18h16" />
            <path d="M8 4v4" />
            <path d="M16 10v4" />
            <path d="M10 16v4" />
          </svg>
        </Link>
      </div>

      <div className="flex flex-col space-y-2">
        {jars.map((jar, index) => {
          const style = jarStyles[index % jarStyles.length];

          const progress = Math.min(
            Math.max(jar.used_percentage, 0),
            100,
          );

          return (
            <div
              key={jar.id}
              className={`flex flex-col gap-2 rounded-2xl border border-slate-100 bg-white p-3.5 shadow-sm transition-colors ${style.border}`}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div
                    className={`flex h-8 w-8 items-center justify-center rounded-xl font-bold ${style.icon}`}
                  >
                    <JarIcon>{style.iconPath}</JarIcon>
                  </div>

                  <div>
                    <span className="block text-xs font-bold text-slate-900">
                      {jar.name}
                    </span>

                    <span className="text-[10px] text-slate-400">
                      Available {formatMoney(jar.available)}
                    </span>
                  </div>
                </div>

                <div className="text-right">
                  <span
                    className={`rounded-full px-2 py-0.5 text-xs font-extrabold ${style.badge}`}
                  >
                    {getAllocationLabel(jar)}
                  </span>
                </div>
              </div>

              <div className="h-2 w-full overflow-hidden rounded-full bg-slate-100 p-0.5">
                <div
                  className={`h-full rounded-full transition-all duration-500 ${style.progress}`}
                  style={{ width: `${progress}%` }}
                />
              </div>

              <div className="flex items-center justify-between text-[11px]">
                <span className="font-extrabold text-slate-900">
                  {formatMoney(jar.used)}
                </span>

                <span className="font-bold text-slate-500">
                  {formatPercentage(progress)} Used
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </section>
  );
}
