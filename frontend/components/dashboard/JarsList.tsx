"use client";

import { useRef, useState } from "react";
import { ChevronRight } from "@material-symbols-svg/react/w400";

import Link from "next/link";
import { useRouter } from "next/navigation";

import { formatMoney } from "@/lib/utils";
import { getJarIcon } from "@/components/jars/jar-icons";

interface Jar {
  id: string;
  name: string;
  icon_key: string;
  allocation_type: "fixed" | "percentage" | "remainder";
  allocation_value: number;
  carry_in: number;
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
  "#F43F5E",
  "#6366F1",
  "#07B682",
  "#F6A723",
];

function getAllocationLabel(jar: Jar) {
  switch (jar.allocation_type) {
    case "percentage":
      return `${jar.allocation_value}% of income`;

    case "fixed":
      return `₹${formatMoney(jar.allocation_value)} fixed`;

    case "remainder":
      return "Remainder";
  }
}

function getStatus(jar: Jar) {
  if (jar.available <= 0) {
    return {
      label: "Fully Spent",
      className:
        "rounded-full border border-rose-200 bg-rose-50 px-1.5 py-0.5 text-[10px] font-semibold text-rose-700",
    };
  }

  return {
    label: `${Math.round(jar.available_percentage)}% Left`,
    className:
      "rounded-full border border-emerald-200 bg-emerald-50 px-1.5 py-0.5 text-[10px] font-semibold text-emerald-700",
  };
}

export default function JarsList({ jars }: JarsListProps) {
  const router = useRouter();
  const [selectedJarId, setSelectedJarId] = useState<string | null>(null);
  const longPressTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const longPressTriggered = useRef(false);

  if (!jars || jars.length === 0) return null;

  function toggleJar(jar: Jar) {
    setSelectedJarId((currentId) => currentId === jar.id ? null : jar.id);
  }

  function startLongPress(jar: Jar, button: number) {
    if (button !== 0) return;
    if (longPressTimer.current) clearTimeout(longPressTimer.current);
    longPressTriggered.current = false;
    longPressTimer.current = setTimeout(() => {
      longPressTriggered.current = true;
      router.push(`/transactions?jar_id=${encodeURIComponent(jar.id)}`);
    }, 600);
  }

  function cancelLongPress() {
    if (longPressTimer.current) {
      clearTimeout(longPressTimer.current);
      longPressTimer.current = null;
    }
  }

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

        <Link
          href="/jars"
          className="flex items-center gap-0.5 text-xs font-semibold text-primary hover:underline"
        >
          View All ({jars.length})
          <ChevronRight size={14} color="currentColor" />
        </Link>
      </div>

      <div className="flex flex-col gap-2">
        {jars.map((jar, index) => {
          const isSelected = selectedJarId === jar.id;
          const displayAllocated = isSelected ? jar.allocated + jar.carry_in : jar.allocated;
          const monthlyAvailable = jar.allocated - jar.used;
          const available = isSelected ? jar.available : monthlyAvailable;
          const usedPercentage = displayAllocated > 0
            ? (jar.used / displayAllocated) * 100
            : 0;
          const displayJar = {
            ...jar,
            allocated: displayAllocated,
            available,
            used_percentage: usedPercentage,
            available_percentage: displayAllocated > 0
              ? (available / displayAllocated) * 100
              : 0,
          };
          const Icon = getJarIcon(jar.icon_key);
          const color = JAR_COLORS[index % JAR_COLORS.length];
          const status = getStatus(displayJar);

          return (
            <button
              key={jar.id}
              type="button"
              onPointerDown={(event) => startLongPress(jar, event.button)}
              onPointerUp={cancelLongPress}
              onPointerLeave={cancelLongPress}
              onPointerCancel={cancelLongPress}
              onClick={() => {
                if (longPressTriggered.current) {
                  longPressTriggered.current = false;
                  return;
                }
                toggleJar(jar);
              }}
              aria-pressed={selectedJarId === jar.id}
              className="flex w-full flex-col gap-2 rounded-xl bg-white p-3 text-left shadow-sm transition-transform active:scale-[0.99]"
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
                  {selectedJarId === jar.id ? (
                    <>
                      ₹{formatMoney(jar.carry_in)}
                      <span>carry in ·</span>
                      ₹{formatMoney(displayAllocated)}
                      <span>allocated</span>
                    </>
                  ) : (
                    <>
                      ₹{formatMoney(displayJar.used)}
                      <span>spent of</span>
                      ₹{formatMoney(displayJar.allocated)}
                    </>
                  )}
                </span>

                <div className="h-1.5 w-full overflow-hidden rounded-full bg-slate-100">
                  <div
                    className="h-full rounded-full transition-all"
                    style={{
                      width: `${Math.min(displayJar.used_percentage, 100)}%`,
                      backgroundColor: color,
                    }}
                  />
                </div>
              </div>

              {/* Bottom */}
              <div className="flex items-center justify-between">
                <span className="flex items-center gap-1 text-xs font-bold text-on-surface">
                  ₹{formatMoney(displayJar.available)}
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
