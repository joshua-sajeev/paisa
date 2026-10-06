"use client";

import { useEffect, useState } from "react";
import {
  Movie,
  Savings,
  ReceiptLong,
  TrendingUp,
  VolunteerActivism,
} from "@material-symbols-svg/react/w400";

import { apiFetch } from "@/lib/api";
import type { DateRange } from "@/components/ui/DateRangePicker";

type AllocationType = "percentage" | "fixed" | "remainder";

type Jar = {
  id: string;
  name: string;
  icon_key: string;
  allocation_type: AllocationType;
  allocation_value: number;
  allocated_amount: number;
  used_amount: number;
  available_amount: number;
  used_percentage: number;
  available_percentage: number;
  is_archived: boolean;
  created_at: string;
  updated_at: string;
};

type Props = {
  dateRange: DateRange;
};

const JAR_COLORS = [
  {
    accent: "#059669", // emerald-600
    border: "border-slate-200 hover:border-emerald-300",
    badge: "bg-emerald-50 text-emerald-700 border-emerald-200/60",
    lightBg: "bg-emerald-500/10",
  },
  {
    accent: "#4F46E5", // indigo-600
    border: "border-slate-200 hover:border-indigo-300",
    badge: "bg-indigo-50 text-indigo-700 border-indigo-200/60",
    lightBg: "bg-indigo-500/10",
  },
  {
    accent: "#D97706", // amber-600
    border: "border-slate-200 hover:border-amber-300",
    badge: "bg-amber-50 text-amber-700 border-amber-200/60",
    lightBg: "bg-amber-500/10",
  },
  {
    accent: "#DB2777", // pink-600
    border: "border-slate-200 hover:border-pink-300",
    badge: "bg-pink-50 text-pink-700 border-pink-200/60",
    lightBg: "bg-pink-500/10",
  },
  {
    accent: "#7C3AED", // violet-600
    border: "border-slate-200 hover:border-violet-300",
    badge: "bg-violet-50 text-violet-700 border-violet-200/60",
    lightBg: "bg-violet-500/10",
  },
];

const JAR_ICONS = {
  jar: Savings,
  necessities: ReceiptLong,
  leisure: Movie,
  giving: VolunteerActivism,
  investment: TrendingUp,
} as const;

function formatAmount(paise: number) {
  return new Intl.NumberFormat("en-IN", {
    style: "currency",
    currency: "INR",
    maximumFractionDigits: 0,
  }).format(paise / 100);
}

function formatAllocation(jar: Jar) {
  switch (jar.allocation_type) {
    case "percentage":
      return `${jar.allocation_value}% rule`;

    case "fixed":
      return `${formatAmount(jar.allocation_value)} fixed`;

    case "remainder":
      return "Remainder sweep";

    default:
      return "";
  }
}

function getJarIcon(iconKey: string) {
  return JAR_ICONS[iconKey as keyof typeof JAR_ICONS] ?? Savings;
}

function getJarColor(index: number) {
  return JAR_COLORS[index % JAR_COLORS.length];
}

export default function JarsList({ dateRange }: Props) {
  const [jars, setJars] = useState<Jar[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;

    async function loadJars() {
      const params = new URLSearchParams();

      if (dateRange?.start) {
        params.set("start_date", dateRange.start);
      }

      if (dateRange?.end) {
        params.set("end_date", dateRange.end);
      }

      const query = params.toString();

      try {
        const data = await apiFetch<Jar[]>(
          `/api/v1/jars${query ? `?${query}` : ""}`,
        );

        if (cancelled) return;

        setJars(data.filter((jar) => !jar.is_archived));
      } catch (error) {
        console.error("Failed to load jars:", error);

        if (!cancelled) {
          setJars([]);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    loadJars();

    return () => {
      cancelled = true;
    };
  }, [dateRange]);

  if (loading) {
    return (
      <div className="space-y-3">
        {[1, 2, 3].map((index) => (
          <div
            key={index}
            className="h-28 animate-pulse rounded-xl border border-slate-200 bg-slate-50/50"
          />
        ))}
      </div>
    );
  }

  if (jars.length === 0) {
    return (
      <div className="rounded-xl border border-dashed border-slate-200 bg-white px-4 py-8 text-center shadow-2xs">
        <p className="text-xs font-semibold text-slate-800">
          No active jars
        </p>

        <p className="mt-1 text-[11px] text-slate-500">
          Create a jar to start allocating money.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {jars.map((jar, index) => {
        const color = getJarColor(index);
        const Icon = getJarIcon(jar.icon_key);

        const usedPercentage = Math.min(
          Math.max(jar.used_percentage, 0),
          100,
        );

        return (
          <div
            key={jar.id}
            className={`group relative cursor-pointer overflow-hidden rounded-xl border bg-white p-3.5 shadow-2xs transition-all duration-200 hover:shadow-sm ${color.border}`}
          >
            {/* Top Row */}
            <div className="flex items-center justify-between gap-3">
              <div className="flex min-w-0 items-center gap-3">
                <div
                  className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-white shadow-2xs transition-transform duration-200 group-hover:scale-105"
                  style={{ backgroundColor: color.accent }}
                >
                  <Icon size={18} />
                </div>

                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <h3 className="truncate text-xs font-bold text-slate-900 group-hover:text-slate-950">
                      {jar.name}
                    </h3>

                    <span
                      className={`inline-flex items-center rounded-md border px-1.5 py-0.5 text-[10px] font-medium tracking-wide ${color.badge}`}
                    >
                      {formatAllocation(jar)}
                    </span>
                  </div>
                </div>
              </div>

              <div className="shrink-0 text-right">
                <div className="text-xs font-bold tabular-nums text-slate-900">
                  {formatAmount(jar.available_amount)}
                </div>

                <div className="text-[10px] font-medium text-slate-400">
                  available
                </div>
              </div>
            </div>

            {/* Bottom Row */}
            <div className="mt-3.5 space-y-1.5 border-t border-slate-100 pt-2.5">
              <div className="flex items-center justify-between text-[11px]">
                <span className="font-normal text-slate-500">
                  Allocated:{" "}
                  <strong className="font-semibold tabular-nums text-slate-700">
                    {formatAmount(jar.allocated_amount)}
                  </strong>
                </span>

                <span
                  className="text-[10px] font-semibold tabular-nums"
                  style={{ color: color.accent }}
                >
                  {jar.used_percentage.toFixed(0)}% reached
                </span>
              </div>

              <div className="flex h-1.5 w-full items-center overflow-hidden rounded-full bg-slate-100">
                <div
                  className="h-full rounded-full transition-all duration-300"
                  style={{
                    width: `${usedPercentage}%`,
                    backgroundColor: color.accent,
                  }}
                />
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
}
