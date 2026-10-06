"use client";

import { useCallback, useEffect, useState } from "react";
import {
  Edit,
  Archive,
  ExpandCircleDown,
  Add,
} from "@material-symbols-svg/react/w400";

import { apiFetch } from "@/lib/api";
import type { DateRange } from "@/components/ui/DateRangePicker";
import AddJarModal from "@/components/jars/AddJarModal";
import EditJarModal from "@/components/jars/EditJarModal";
import { EditAllocationsModal } from "@/components/jars/EditAllocationsModal";
import { getJarIcon } from "@/components/jars/jar-icons";

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
    accent: "#059669",
    border: "border-slate-200 hover:border-emerald-300",
    badge: "bg-emerald-50 text-emerald-700 border-emerald-200/60",
  },
  {
    accent: "#4F46E5",
    border: "border-slate-200 hover:border-indigo-300",
    badge: "bg-indigo-50 text-indigo-700 border-indigo-200/60",
  },
  {
    accent: "#D97706",
    border: "border-slate-200 hover:border-amber-300",
    badge: "bg-amber-50 text-amber-700 border-amber-200/60",
  },
  {
    accent: "#DB2777",
    border: "border-slate-200 hover:border-pink-300",
    badge: "bg-pink-50 text-pink-700 border-pink-200/60",
  },
  {
    accent: "#7C3AED",
    border: "border-slate-200 hover:border-violet-300",
    badge: "bg-violet-50 text-violet-700 border-violet-200/60",
  },
];

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

function getJarColor(index: number) {
  return JAR_COLORS[index % JAR_COLORS.length];
}

export default function JarsList({ dateRange }: Props) {
  const [jars, setJars] = useState<Jar[]>([]);
  const [showArchived, setShowArchived] = useState(false);
  const [showAddJar, setShowAddJar] = useState(false);

  const [showEditModal, setShowEditModal] = useState(false);
  const [showEditAllocationsModal, setShowEditAllocationsModal] =
    useState(false);
  const [showArchiveModal, setShowArchiveModal] = useState(false);

  const [selectedJar, setSelectedJar] = useState<Jar | null>(null);

  const loadJars = useCallback(
    async (signal?: AbortSignal) => {
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
          signal ? { signal } : undefined,
        );

        setJars(data);
      } catch (error) {
        if (signal?.aborted) return;

        console.error("Failed to load jars:", error);
        setJars([]);
      }
    },
    [dateRange],
  );

  /*
   * Keep the fetch inside the async callback.
   *
   * Do not call loadJars() directly from this effect because
   * react-hooks/set-state-in-effect flags that pattern.
   */
  useEffect(() => {
    const controller = new AbortController();

    const params = new URLSearchParams();

    if (dateRange?.start) {
      params.set("start_date", dateRange.start);
    }

    if (dateRange?.end) {
      params.set("end_date", dateRange.end);
    }

    const query = params.toString();

    async function fetchJars() {
      try {
        const data = await apiFetch<Jar[]>(
          `/api/v1/jars${query ? `?${query}` : ""}`,
          {
            signal: controller.signal,
          },
        );

        console.log("JARS API RESPONSE:", data);
        setJars(data);
      } catch (error) {
        if (controller.signal.aborted) return;

        console.error("Failed to load jars:", error);
        setJars([]);
      }
    }

    void fetchJars();

    return () => {
      controller.abort();
    };
  }, [dateRange]);

  const updateArchiveStatus = async (
    jar: Jar,
    isArchived: boolean,
  ) => {
    try {
      await apiFetch(`/api/v1/jars/${jar.id}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          is_archived: isArchived,
        }),
      });

      setShowArchiveModal(false);
      setSelectedJar(null);

      await loadJars();
    } catch (error) {
      console.error(
        `Failed to ${isArchived ? "archive" : "unarchive"} jar:`,
        error,
      );
    }
  };

  const activeJars = jars.filter((jar) => !jar.is_archived);
  const archivedJars = jars.filter((jar) => jar.is_archived);

  const renderJar = (jar: Jar, index: number) => {
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
                <h3 className="truncate text-xs font-bold text-slate-900">
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

        <div className="mt-3 flex items-center justify-between border-t border-slate-100 pt-2.5">
          <div className="flex items-center gap-1">
            <button
              type="button"
              onClick={(event) => {
                event.stopPropagation();
                setSelectedJar(jar);
                setShowEditModal(true);
              }}
              className="flex items-center gap-1 rounded-lg px-2 py-1 text-xs font-semibold text-slate-500 transition hover:bg-slate-50 hover:text-slate-800"
            >
              <Edit size={15} color="currentColor" />
              <span>Edit</span>
            </button>

            <button
              type="button"
              onClick={(event) => {
                event.stopPropagation();
                setSelectedJar(jar);
                setShowArchiveModal(true);
              }}
              className="flex items-center gap-1 rounded-lg px-2 py-1 text-xs font-semibold text-slate-500 transition hover:bg-slate-50 hover:text-slate-800"
            >
              <Archive size={15} color="currentColor" />
              <span>
                {jar.is_archived ? "Unarchive" : "Archive"}
              </span>
            </button>
          </div>
        </div>
      </div>
    );
  };

  return (
    <>
      <div className="space-y-6">
        {activeJars.length === 0 && archivedJars.length === 0 && (
          <div className="rounded-xl border border-dashed border-slate-200 bg-white px-4 py-8 text-center shadow-2xs">
            <p className="text-xs font-semibold text-slate-800">
              No jars
            </p>

            <p className="mt-1 text-[11px] text-slate-500">
              Create a jar to start allocating money.
            </p>

            <button
              type="button"
              onClick={() => setShowAddJar(true)}
              className="mt-4 inline-flex items-center gap-1.5 rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white shadow-2xs transition-colors hover:bg-slate-800"
            >
              <Add size={14} color="currentColor" />
              Add Jar
            </button>
          </div>
        )}

        {activeJars.length > 0 && (
          <section className="space-y-3">
            <div className="flex items-center justify-between gap-3">
              <h2 className="text-[15px] font-bold tracking-tight text-slate-900">
                Active Jars
              </h2>

              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={() => setShowEditAllocationsModal(true)}
                  className="inline-flex items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-xs font-semibold text-slate-600 shadow-2xs transition-colors hover:bg-slate-50 hover:text-slate-800"
                >
                  <Edit size={14} color="currentColor" />
                  Edit Allocations
                </button>

                <button
                  type="button"
                  onClick={() => setShowAddJar(true)}
                  className="inline-flex items-center gap-1.5 rounded-lg bg-slate-900 px-2.5 py-1.5 text-xs font-semibold text-white shadow-2xs transition-colors hover:bg-slate-800"
                >
                  <Add size={14} color="currentColor" />
                  Add Jar
                </button>
              </div>
            </div>

            <div className="space-y-3">
              {activeJars.map((jar, index) =>
                renderJar(jar, index),
              )}
            </div>
          </section>
        )}

        {archivedJars.length > 0 && (
          <section className="space-y-3">
            <button
              type="button"
              onClick={() => setShowArchived((value) => !value)}
              className="flex w-full items-center justify-between"
            >
              <div className="text-left">
                <h2 className="text-[15px] font-bold tracking-tight text-slate-900">
                  Archived Jars
                </h2>

                <p className="mt-0.5 text-[11px] font-medium text-slate-500">
                  {archivedJars.length} archived
                </p>
              </div>

              <ExpandCircleDown
                width={20}
                height={20}
                className={`shrink-0 text-slate-500 transition-transform duration-200 ${showArchived ? "" : "-rotate-90"
                  }`}
              />
            </button>

            {showArchived && (
              <div className="space-y-3 opacity-75">
                {archivedJars.map((jar, index) =>
                  renderJar(jar, index),
                )}
              </div>
            )}
          </section>
        )}
      </div>

      <AddJarModal
        open={showAddJar}
        onCloseAction={() => setShowAddJar(false)}
        onCreatedAction={loadJars}
      />

      {/* 
       * Only mount this modal when it is actually opened.
       *
       * This is important because EditAllocationsModal initializes
       * its local state from initialJars. If it stays mounted while
       * initialJars is [], its local state remains [] even after
       * the API finishes loading the jars.
       */}
      {showEditAllocationsModal && (
        <EditAllocationsModal
          isOpen
          onClose={() => setShowEditAllocationsModal(false)}
          initialJars={activeJars.map((jar, index) => {
            const color = getJarColor(index);

            return {
              id: jar.id,
              name: jar.name,
              subtitle: formatAllocation(jar),
              iconKey: jar.icon_key,
              iconBgColor: "bg-slate-50",
              iconTextColor: "text-slate-700",
              badgeText: jar.allocation_type.toUpperCase(),
              badgeColor: color.badge,
              allocationType: jar.allocation_type,
              allocationValue: jar.allocation_value,
            };
          })}
          onSave={async (updatedJars) => {
            await apiFetch("/api/v1/jars/allocations", {
              method: "PATCH",
              headers: {
                "Content-Type": "application/json",
              },
              body: JSON.stringify({
                allocations: updatedJars.map((jar) => ({
                  id: jar.id,
                  allocation_type: jar.allocationType,
                  allocation_value: jar.allocationValue,
                })),
              }),
            });

            await loadJars();
          }}
        />
      )}

      {selectedJar && showEditModal && (
        <EditJarModal
          key={selectedJar.id}
          open={showEditModal}
          jar={selectedJar}
          onCloseAction={() => {
            setShowEditModal(false);
            setSelectedJar(null);
          }}
          onUpdatedAction={loadJars}
        />
      )}

      {showArchiveModal && selectedJar && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 px-4"
          onClick={() => {
            setShowArchiveModal(false);
            setSelectedJar(null);
          }}
        >
          <div
            className="w-full max-w-sm rounded-2xl bg-white p-5 shadow-xl"
            onClick={(event) => event.stopPropagation()}
          >
            <h2 className="text-sm font-bold text-slate-900">
              {selectedJar.is_archived
                ? "Unarchive jar?"
                : "Archive jar?"}
            </h2>

            <p className="mt-1.5 text-xs leading-5 text-slate-500">
              Are you sure you want to{" "}
              {selectedJar.is_archived ? "unarchive" : "archive"}{" "}
              <span className="font-semibold text-slate-700">
                {selectedJar.name}
              </span>
              ?
            </p>

            <div className="mt-5 flex justify-end gap-2">
              <button
                type="button"
                onClick={() => {
                  setShowArchiveModal(false);
                  setSelectedJar(null);
                }}
                className="rounded-lg px-3 py-2 text-xs font-semibold text-slate-600 transition hover:bg-slate-50"
              >
                Cancel
              </button>

              <button
                type="button"
                onClick={() => {
                  void updateArchiveStatus(
                    selectedJar,
                    !selectedJar.is_archived,
                  );
                }}
                className="rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white transition hover:bg-slate-800"
              >
                {selectedJar.is_archived ? "Unarchive" : "Archive"}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
