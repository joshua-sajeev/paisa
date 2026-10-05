"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  Search,
  Download,
  ChevronLeft,
  ChevronRight,
  FilterList,
  ExpandCircleDown,
} from "@material-symbols-svg/react/w400";

import {
  Allocation,
  getAllocations,
} from "@/lib/jars";
import { formatMonth, formatMoney } from "@/lib/utils";
import { AllocationRow } from "./AllocationRow";

const PAGE_SIZE = 4;

type Props = {
  month?: string;
};

const jarColors = [
  "#6366F1",
  "#07B682",
  "#F59E0B",
  "#EC4899",
  "#8B5CF6",
] as const;

export default function AllocationHistory({ month }: Props) {
  const [allocations, setAllocations] = useState<Allocation[]>([]);
  const [total, setTotal] = useState(0);

  const [summary, setSummary] = useState({
    month: month ?? "",
    total_allocated: 0,
  });

  const [search, setSearch] = useState("");
  const [appliedSearch, setAppliedSearch] = useState("");

  const [jarFilter, setJarFilter] = useState("");
  const [minAmount, setMinAmount] = useState("");
  const [maxAmount, setMaxAmount] = useState("");

  const [filterPanelOpen, setFilterPanelOpen] = useState(false);
  const [jarOpen, setJarOpen] = useState(false);

  const filterPanelRef = useRef<HTMLDivElement>(null);

  const [page, setPage] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  /*
   * Keep jar identity separate from the visible name.
   * This means renamed jars won't break the filter.
   */
  const jars = useMemo(() => {
    const map = new Map<string, string>();

    allocations.forEach((allocation) => {
      if (!map.has(allocation.jar_id)) {
        map.set(allocation.jar_id, allocation.jar_name);
      }
    });

    return Array.from(map.entries())
      .map(([id, name]) => ({ id, name }))
      .sort((a, b) => a.name.localeCompare(b.name));
  }, [allocations]);

  /*
   * Assign a consistent color to each jar ID.
   */
  const jarColorMap = useMemo(() => {
    const map = new Map<string, string>();

    allocations.forEach((allocation) => {
      if (!map.has(allocation.jar_id)) {
        map.set(
          allocation.jar_id,
          jarColors[map.size % jarColors.length],
        );
      }
    });

    return map;
  }, [allocations]);

  const activeFiltersCount =
    (jarFilter ? 1 : 0) +
    (minAmount ? 1 : 0) +
    (maxAmount ? 1 : 0);

  const hasActiveFilters = activeFiltersCount > 0;

  const loadAllocations = useCallback(
    async (signal?: AbortSignal) => {
      try {
        setLoading(true);
        setError(null);

        /*
         * UI amounts are entered in rupees.
         * API amounts are sent in paise.
         */
        const parsedMinAmount = minAmount
          ? Number(minAmount)
          : undefined;

        const parsedMaxAmount = maxAmount
          ? Number(maxAmount)
          : undefined;

        const result = await getAllocations({
          month,
          search: appliedSearch || undefined,
          jarId: jarFilter || undefined,

          minAmount:
            parsedMinAmount !== undefined &&
              Number.isFinite(parsedMinAmount)
              ? Math.round(parsedMinAmount * 100)
              : undefined,

          maxAmount:
            parsedMaxAmount !== undefined &&
              Number.isFinite(parsedMaxAmount)
              ? Math.round(parsedMaxAmount * 100)
              : undefined,

          limit: PAGE_SIZE,
          offset: page * PAGE_SIZE,
        });

        if (signal?.aborted) return;

        setAllocations(result.allocations);
        setTotal(result.total);
        setSummary(result.monthly_summary);
      } catch (err) {
        if (signal?.aborted) return;

        console.error(err);
        setError("Failed to load allocation history.");
      } finally {
        if (!signal?.aborted) {
          setLoading(false);
        }
      }
    },
    [
      month,
      appliedSearch,
      jarFilter,
      minAmount,
      maxAmount,
      page,
    ],
  );

  useEffect(() => {
    const controller = new AbortController();

    // eslint-disable-next-line react-hooks/set-state-in-effect
    loadAllocations(controller.signal);

    return () => {
      controller.abort();
    };
  }, [loadAllocations]);

  useEffect(() => {
    function handleOutside(event: MouseEvent | TouchEvent) {
      if (
        filterPanelRef.current &&
        !filterPanelRef.current.contains(event.target as Node)
      ) {
        setFilterPanelOpen(false);
        setJarOpen(false);
      }
    }

    document.addEventListener("mousedown", handleOutside);
    document.addEventListener("touchstart", handleOutside);

    return () => {
      document.removeEventListener("mousedown", handleOutside);
      document.removeEventListener("touchstart", handleOutside);
    };
  }, []);

  const totalPages = Math.ceil(total / PAGE_SIZE);

  const showingFrom =
    total === 0 ? 0 : page * PAGE_SIZE + 1;

  const showingTo = Math.min(
    (page + 1) * PAGE_SIZE,
    total,
  );

  const monthLabel = useMemo(() => {
    if (!summary.month) {
      return "";
    }

    return formatMonth(summary.month);
  }, [summary.month]);

  function handleSearchSubmit(
    event: React.SyntheticEvent<HTMLFormElement>,
  ) {
    event.preventDefault();

    setPage(0);
    setAppliedSearch(search.trim());
  }

  function handleJarChange(value: string) {
    setJarFilter(value);
    setPage(0);
    setJarOpen(false);
  }

  function handleMinAmountChange(value: string) {
    setMinAmount(value);
    setPage(0);
  }

  function handleMaxAmountChange(value: string) {
    setMaxAmount(value);
    setPage(0);
  }

  function resetFilters() {
    setJarFilter("");
    setMinAmount("");
    setMaxAmount("");
    setPage(0);
    setJarOpen(false);
  }

  function previousPage() {
    setPage((current) => Math.max(0, current - 1));
  }

  function nextPage() {
    setPage((current) =>
      Math.min(totalPages - 1, current + 1),
    );
  }

  return (
    <section className="space-y-3">
      {/* Monthly Summary */}
      <div className="relative overflow-hidden rounded-2xl border border-slate-100 bg-white p-3.5 shadow-sm">
        <div className="flex items-center justify-between gap-3">
          <div>
            <div className="mb-1 inline-flex items-center gap-1 rounded-full bg-blue-50 px-2 py-0.5 text-[9px] font-extrabold uppercase tracking-wide text-blue-600">
              <span className="h-1.5 w-1.5 rounded-full bg-blue-600" />
              Monthly Summary
            </div>

            <h2 className="text-sm font-black tracking-tight text-slate-900">
              {monthLabel} Activity
            </h2>

            <p className="mt-0.5 text-[11px] font-medium text-slate-500">
              {total} allocation{total === 1 ? "" : "s"} this month
            </p>
          </div>

          <div className="shrink-0 rounded-xl border border-slate-100 bg-slate-50 px-3 py-2 text-right">
            <span className="block text-[9px] font-bold uppercase tracking-wider text-slate-400">
              Total Allocated
            </span>

            <span className="text-base font-black tracking-tight text-[#07B682]">
              ₹{formatMoney(summary.total_allocated)}
            </span>
          </div>
        </div>
      </div>

      {/* Allocation History */}
      <div className="overflow-hidden rounded-2xl border border-slate-100 bg-white shadow-sm">
        {/* Header */}
        <div className="flex items-center justify-between gap-2 border-b border-slate-100 bg-slate-50/80 px-3.5 py-3">
          <div className="flex items-center gap-2">
            <span className="text-xs font-bold tracking-tight text-slate-800">
              Allocation History
            </span>

            <span className="rounded-full border border-indigo-100 bg-blue-50 px-2 py-0.5 text-[10px] font-bold text-blue-600">
              {total}
            </span>
          </div>

          <span className="text-[11px] font-semibold text-slate-500">
            Newest first
          </span>
        </div>

        {/* Toolbar */}
        <div className="border-b border-slate-100 bg-white p-3">
          <div className="flex items-center gap-2">
            {/* Search */}
            <form
              onSubmit={handleSearchSubmit}
              className="relative flex-1"
            >
              <Search className="pointer-events-none absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-slate-400" />

              <input
                value={search}
                onChange={(event) =>
                  setSearch(event.target.value)
                }
                type="text"
                placeholder="Search transaction..."
                className="h-[34px] w-full rounded-xl border border-slate-200 bg-slate-50 py-2 pl-9 pr-3 text-xs font-medium text-slate-800 placeholder-slate-400 outline-none transition-colors focus:border-blue-600"
              />
            </form>

            {/* Filter Button */}
            <div
              ref={filterPanelRef}
              className="relative shrink-0"
            >
              <button
                type="button"
                onClick={() =>
                  setFilterPanelOpen((previous) => !previous)
                }
                className={`flex h-[34px] items-center gap-2 rounded-xl border px-3 text-xs font-bold transition ${hasActiveFilters
                    ? "border-slate-900 bg-slate-900 text-white"
                    : "border-slate-200 bg-white text-slate-700 hover:border-slate-300"
                  }`}
              >
                <FilterList className="h-3.5 w-3.5" />

                <span>Filters</span>

                {hasActiveFilters && (
                  <span className="flex h-4.5 min-w-4.5 items-center justify-center rounded-full bg-white px-1 text-[9px] font-extrabold text-slate-900">
                    {activeFiltersCount}
                  </span>
                )}

                <ExpandCircleDown
                  className={`h-3.5 w-3.5 transition-transform ${filterPanelOpen ? "" : "-rotate-90"
                    }`}
                />
              </button>

              {/* Filter Panel */}
              {filterPanelOpen && (
                <div className="absolute right-0 z-30 mt-2 w-80 rounded-2xl border border-slate-200 bg-white p-4 shadow-xl">
                  {/* Panel Header */}
                  <div className="mb-4">
                    <p className="text-xs font-black text-slate-900">
                      Filter allocations
                    </p>

                    <p className="mt-0.5 text-[10px] font-medium text-slate-400">
                      Narrow results by jar or amount range.
                    </p>
                  </div>

                  {/* Jar */}
                  <div className="relative">
                    <label className="mb-1.5 block text-[10px] font-extrabold uppercase tracking-wider text-slate-400">
                      Jar
                    </label>

                    <button
                      type="button"
                      onClick={() =>
                        setJarOpen((previous) => !previous)
                      }
                      className={`flex h-10 w-full items-center justify-between rounded-xl border bg-white px-3 text-xs font-bold outline-none transition ${jarOpen
                          ? "border-blue-500 ring-2 ring-blue-500/10"
                          : "border-slate-200 hover:border-slate-300"
                        }`}
                    >
                      <span
                        className={
                          jarFilter
                            ? "truncate text-slate-800"
                            : "text-slate-400"
                        }
                      >
                        {jarFilter
                          ? jars.find(
                            (jar) => jar.id === jarFilter,
                          )?.name ?? "Selected jar"
                          : "All jars"}
                      </span>

                      <ExpandCircleDown
                        className={`h-4 w-4 shrink-0 text-slate-400 transition-transform ${jarOpen ? "" : "-rotate-90"
                          }`}
                      />
                    </button>

                    {jarOpen && (
                      <div className="absolute left-0 right-0 z-40 mt-1 max-h-48 overflow-y-auto rounded-xl border border-slate-200 bg-white p-1.5 shadow-lg">
                        <button
                          type="button"
                          onClick={() =>
                            handleJarChange("")
                          }
                          className={`w-full rounded-lg px-3 py-2 text-left text-xs font-bold transition ${!jarFilter
                              ? "bg-slate-900 text-white"
                              : "text-slate-700 hover:bg-slate-50"
                            }`}
                        >
                          All jars
                        </button>

                        {jars.map((jar) => {
                          const color =
                            jarColorMap.get(jar.id) ??
                            jarColors[0];

                          const selected =
                            jarFilter === jar.id;

                          return (
                            <button
                              key={jar.id}
                              type="button"
                              onClick={() =>
                                handleJarChange(jar.id)
                              }
                              className={`flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-xs font-bold transition ${selected
                                  ? "bg-slate-900 text-white"
                                  : "text-slate-700 hover:bg-slate-50"
                                }`}
                            >
                              <span
                                className="h-2 w-2 shrink-0 rounded-full"
                                style={{
                                  backgroundColor: selected
                                    ? "white"
                                    : color,
                                }}
                              />

                              <span className="truncate">
                                {jar.name}
                              </span>
                            </button>
                          );
                        })}
                      </div>
                    )}
                  </div>

                  {/* Amount Range */}
                  <div className="mt-4">
                    <label className="mb-1.5 block text-[10px] font-extrabold uppercase tracking-wider text-slate-400">
                      Amount range
                    </label>

                    <div className="grid grid-cols-2 gap-2">
                      {/* Minimum */}
                      <div className="relative">
                        <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-xs font-bold text-slate-400">
                          ₹
                        </span>

                        <input
                          type="number"
                          min="0"
                          value={minAmount}
                          onChange={(event) =>
                            handleMinAmountChange(
                              event.target.value,
                            )
                          }
                          placeholder="Min"
                          className="h-10 w-full rounded-xl border border-slate-200 bg-slate-50 pl-7 pr-3 text-xs font-bold text-slate-800 outline-none transition focus:border-blue-500 focus:bg-white focus:ring-2 focus:ring-blue-500/10 placeholder:text-slate-400"
                        />
                      </div>

                      {/* Maximum */}
                      <div className="relative">
                        <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-xs font-bold text-slate-400">
                          ₹
                        </span>

                        <input
                          type="number"
                          min="0"
                          value={maxAmount}
                          onChange={(event) =>
                            handleMaxAmountChange(
                              event.target.value,
                            )
                          }
                          placeholder="Max"
                          className="h-10 w-full rounded-xl border border-slate-200 bg-slate-50 pl-7 pr-3 text-xs font-bold text-slate-800 outline-none transition focus:border-blue-500 focus:bg-white focus:ring-2 focus:ring-blue-500/10 placeholder:text-slate-400"
                        />
                      </div>
                    </div>
                  </div>

                  {/* Actions */}
                  <div className="mt-4 flex items-center justify-between border-t border-slate-100 pt-3">
                    <button
                      type="button"
                      onClick={resetFilters}
                      disabled={!hasActiveFilters}
                      className="text-xs font-bold text-slate-400 transition hover:text-slate-700 disabled:cursor-default disabled:opacity-40"
                    >
                      Reset
                    </button>

                    <button
                      type="button"
                      onClick={() => {
                        setFilterPanelOpen(false);
                        setJarOpen(false);
                      }}
                      className="rounded-xl bg-slate-900 px-4 py-2 text-xs font-bold text-white transition hover:bg-slate-800"
                    >
                      Apply
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Loading */}
        {loading && (
          <div className="divide-y divide-slate-100">
            {[1, 2, 3, 4].map((item) => (
              <div
                key={item}
                className="flex items-center gap-3 p-3"
              >
                <div className="h-10 w-10 animate-pulse rounded-xl bg-slate-100" />

                <div className="flex-1 space-y-1.5">
                  <div className="h-3 w-28 animate-pulse rounded bg-slate-100" />
                  <div className="h-2.5 w-20 animate-pulse rounded bg-slate-100" />
                </div>

                <div className="space-y-1.5">
                  <div className="ml-auto h-3.5 w-16 animate-pulse rounded bg-slate-100" />
                  <div className="ml-auto h-2.5 w-10 animate-pulse rounded bg-slate-100" />
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Error */}
        {!loading && error && (
          <div className="p-6 text-center">
            <p className="text-xs font-semibold text-red-500">
              {error}
            </p>

            <button
              type="button"
              onClick={() => loadAllocations()}
              className="mt-2.5 rounded-lg bg-blue-600 px-2.5 py-1 text-[11px] font-bold text-white"
            >
              Retry
            </button>
          </div>
        )}

        {/* Empty */}
        {!loading && !error && allocations.length === 0 && (
          <div className="p-6 text-center">
            <div className="mx-auto flex h-10 w-10 items-center justify-center rounded-xl bg-blue-50 text-blue-600">
              <FilterList className="h-5 w-5" />
            </div>

            <h3 className="mt-2.5 text-xs font-extrabold text-slate-900">
              No allocations found
            </h3>

            <p className="mx-auto mt-0.5 max-w-xs text-[11px] leading-relaxed text-slate-500">
              Try changing your search, jar, or amount range.
            </p>
          </div>
        )}

        {/* Rows */}
        {!loading && !error && allocations.length > 0 && (
          <div className="divide-y divide-slate-100">
            {allocations.map((allocation) => (
              <AllocationRow
                key={allocation.id}
                allocation={allocation}
                color={
                  jarColorMap.get(allocation.jar_id) ??
                  jarColors[0]
                }
              />
            ))}
          </div>
        )}

        {/* Footer */}
        {!loading && !error && total > 0 && (
          <div className="flex flex-col gap-2 border-t border-slate-100 bg-slate-50/80 px-3.5 py-2.5">
            <div className="flex items-center justify-between text-[11px]">
              <span className="font-medium text-slate-500">
                Showing {showingFrom}-{showingTo} of {total} allocations
              </span>

              <button
                type="button"
                className="flex items-center gap-1 font-bold text-blue-600 transition-colors hover:text-blue-700"
              >
                <Download className="h-3 w-3" />
                Download
              </button>
            </div>

            <div className="flex items-center justify-between border-t border-slate-200/80 pt-1">
              <button
                type="button"
                onClick={previousPage}
                disabled={page === 0}
                className="flex items-center gap-1 rounded-lg border border-slate-200 bg-white px-2.5 py-1 text-[11px] font-bold text-slate-700 transition-colors hover:bg-slate-50 disabled:cursor-default disabled:text-slate-400"
              >
                <ChevronLeft className="h-3 w-3" />
                Prev
              </button>

              <div className="flex items-center gap-1">
                {Array.from(
                  { length: totalPages },
                  (_, index) => index,
                )
                  .slice(
                    Math.max(0, page - 2),
                    Math.min(totalPages, page + 3),
                  )
                  .map((pageNumber) => (
                    <button
                      key={pageNumber}
                      type="button"
                      onClick={() => setPage(pageNumber)}
                      className={`h-6 w-6 rounded-md text-[11px] font-bold ${pageNumber === page
                          ? "bg-blue-600 text-white shadow-sm"
                          : "text-slate-600 hover:bg-slate-100"
                        }`}
                    >
                      {pageNumber + 1}
                    </button>
                  ))}
              </div>

              <button
                type="button"
                onClick={nextPage}
                disabled={page >= totalPages - 1}
                className="flex items-center gap-1 rounded-lg border border-slate-200 bg-white px-2.5 py-1 text-[11px] font-bold text-slate-700 transition-colors hover:bg-slate-50 disabled:cursor-default disabled:text-slate-400"
              >
                Next
                <ChevronRight className="h-3 w-3" />
              </button>
            </div>
          </div>
        )}
      </div>
    </section>
  );
}
