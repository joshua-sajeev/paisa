"use client";

import type { StatementTransaction } from "./AccountStatement";
import { useEffect, useRef, useState } from "react";
import { Search } from "@material-symbols-svg/react/w400";
import DateRangePicker, {
  type DateRange,
  isInDateRange,
  toDateKey,
} from "@/components/ui/DateRangePicker";

type Filter = "all" | "income" | "expense" | "transfer";

type Props = {
  search: string;
  filter: Filter;
  jarFilter: string;
  categoryFilter: string;
  minAmount: string;
  maxAmount: string;
  dateRange: DateRange;
  transactions: StatementTransaction[];
  onSearchChange: (value: string) => void;
  onFilterChange: (value: Filter) => void;
  onJarFilterChange: (value: string) => void;
  onCategoryFilterChange: (value: string) => void;
  onMinAmountChange: (value: string) => void;
  onMaxAmountChange: (value: string) => void;
  onDateRangeChange: (value: DateRange) => void;
};

const filters: { value: Filter; label: string }[] = [
  { value: "all", label: "All" },
  { value: "income", label: "Income" },
  { value: "expense", label: "Expense" },
  { value: "transfer", label: "Transfers" },
];

export default function StatementFilters({
  search,
  filter,
  jarFilter,
  categoryFilter,
  minAmount,
  maxAmount,
  dateRange,
  transactions,
  onSearchChange,
  onFilterChange,
  onJarFilterChange,
  onCategoryFilterChange,
  onMinAmountChange,
  onMaxAmountChange,
  onDateRangeChange,
}: Props) {
  const [filterPanelOpen, setFilterPanelOpen] = useState(false);

  // Custom dropdown states
  const [categoryOpen, setCategoryOpen] = useState(false);
  const [jarOpen, setJarOpen] = useState(false);

  const filterPanelRef = useRef<HTMLDivElement>(null);

  const jars = Array.from(
    new Set(transactions.map((t) => t.jar_name).filter((jar): jar is string => Boolean(jar))),
  ).sort();

  const categories = Array.from(
    new Set(transactions.map((t) => t.category).filter(Boolean)),
  ).sort();

  const activeFiltersCount =
    (jarFilter !== "all" ? 1 : 0) +
    (categoryFilter !== "all" ? 1 : 0) +
    (minAmount ? 1 : 0) +
    (maxAmount ? 1 : 0);

  const hasActiveAdvancedFilters = activeFiltersCount > 0;

  // Close menus when clicking outside
  useEffect(() => {
    function handleOutside(event: MouseEvent | TouchEvent) {
      if (filterPanelRef.current && !filterPanelRef.current.contains(event.target as Node)) {
        setFilterPanelOpen(false);
      }
    }

    document.addEventListener("mousedown", handleOutside);
    document.addEventListener("touchstart", handleOutside);
    return () => {
      document.removeEventListener("mousedown", handleOutside);
      document.removeEventListener("touchstart", handleOutside);
    };
  }, []);

  return (
    <section className="mb-5 space-y-3">
      {/* Top Type Filters (All, Income, Expense, Transfers) */}
      <div className="flex gap-2 overflow-x-auto pb-1">
        {filters.map((item) => {
          const active = filter === item.value;
          return (
            <button
              key={item.value}
              type="button"
              onClick={() => onFilterChange(item.value)}
              className={`shrink-0 rounded-full px-4 py-2 text-xs font-bold transition ${active
                  ? "bg-black text-white"
                  : "border border-black/10 bg-white text-black/60"
                }`}
            >
              {item.label}
            </button>
          );
        })}
      </div>

      {/* Search and Advanced Filters Toggle */}
      <div className="flex items-center gap-2">
        <div className="flex flex-1 items-center gap-2 rounded-2xl border border-black/10 bg-white px-4">
          <Search className="h-5 w-5 shrink-0 text-black/40" />
          <input
            type="text"
            value={search}
            onChange={(event) => onSearchChange(event.target.value)}
            placeholder="Search name..."
            className="h-12 w-full bg-transparent text-sm font-medium outline-none placeholder:text-black/35"
          />
        </div>

        {/* Filter Toggle Button */}
        <div className="relative" ref={filterPanelRef}>
          <button
            type="button"
            onClick={() => setFilterPanelOpen((prev) => !prev)}
            className={`flex h-12 items-center gap-2 rounded-2xl border px-4 text-xs font-bold transition ${hasActiveAdvancedFilters
                ? "border-black bg-black text-white"
                : "border-black/10 bg-white text-black hover:border-black/30"
              }`}
          >
            <span>Filters</span>
            {hasActiveAdvancedFilters && (
              <span className="flex h-5 w-5 items-center justify-center rounded-full bg-white text-[10px] text-black font-extrabold">
                {activeFiltersCount}
              </span>
            )}
          </button>

          {/* Expanded Filter Panel */}
          {filterPanelOpen && (
            <div className="absolute right-0 z-30 mt-2 w-80 rounded-2xl border border-black/10 bg-white p-4 shadow-xl space-y-4">

              {/* Category Custom Dropdown */}
              <div className="relative">
                <label className="block text-[11px] font-bold uppercase tracking-wider text-black/40 mb-1.5">
                  Category
                </label>
                <button
                  type="button"
                  onClick={() => {
                    setCategoryOpen((prev) => !prev);
                    setJarOpen(false);
                  }}
                  className="flex h-11 w-full items-center justify-between rounded-xl border border-black/10 bg-white px-3 text-xs font-bold text-black outline-none transition hover:border-black/30"
                >
                  <span className="truncate">
                    {categoryFilter === "all" ? "All categories" : categoryFilter}
                  </span>
                  <span className={`text-black/40 transition ${categoryOpen ? "rotate-180" : ""}`}>
                    ⌄
                  </span>
                </button>

                {categoryOpen && (
                  <div className="absolute left-0 right-0 z-40 mt-1 max-h-48 overflow-y-auto rounded-xl border border-black/10 bg-white p-1 shadow-lg space-y-0.5">
                    <button
                      type="button"
                      onClick={() => {
                        onCategoryFilterChange("all");
                        setCategoryOpen(false);
                      }}
                      className={`w-full rounded-lg px-3 py-2 text-left text-xs font-bold transition ${categoryFilter === "all"
                          ? "bg-black text-white"
                          : "text-black/75 hover:bg-black/5 hover:text-black"
                        }`}
                    >
                      All categories
                    </button>
                    {categories.map((cat) => (
                      <button
                        key={cat}
                        type="button"
                        onClick={() => {
                          onCategoryFilterChange(cat);
                          setCategoryOpen(false);
                        }}
                        className={`w-full rounded-lg px-3 py-2 text-left text-xs font-bold transition truncate ${categoryFilter === cat
                            ? "bg-black text-white"
                            : "text-black/75 hover:bg-black/5 hover:text-black"
                          }`}
                      >
                        {cat}
                      </button>
                    ))}
                  </div>
                )}
              </div>

              {/* Jar Custom Dropdown */}
              <div className="relative">
                <label className="block text-[11px] font-bold uppercase tracking-wider text-black/40 mb-1.5">
                  Jars
                </label>
                <button
                  type="button"
                  onClick={() => {
                    setJarOpen((prev) => !prev);
                    setCategoryOpen(false);
                  }}
                  className="flex h-11 w-full items-center justify-between rounded-xl border border-black/10 bg-white px-3 text-xs font-bold text-black outline-none transition hover:border-black/30"
                >
                  <span className="truncate">
                    {jarFilter === "all" ? "All jars" : jarFilter}
                  </span>
                  <span className={`text-black/40 transition ${jarOpen ? "rotate-180" : ""}`}>
                    ⌄
                  </span>
                </button>

                {jarOpen && (
                  <div className="absolute left-0 right-0 z-40 mt-1 max-h-48 overflow-y-auto rounded-xl border border-black/10 bg-white p-1 shadow-lg space-y-0.5">
                    <button
                      type="button"
                      onClick={() => {
                        onJarFilterChange("all");
                        setJarOpen(false);
                      }}
                      className={`w-full rounded-lg px-3 py-2 text-left text-xs font-bold transition ${jarFilter === "all"
                          ? "bg-black text-white"
                          : "text-black/75 hover:bg-black/5 hover:text-black"
                        }`}
                    >
                      All jars
                    </button>
                    {jars.map((jar) => (
                      <button
                        key={jar}
                        type="button"
                        onClick={() => {
                          onJarFilterChange(jar);
                          setJarOpen(false);
                        }}
                        className={`w-full rounded-lg px-3 py-2 text-left text-xs font-bold transition truncate ${jarFilter === jar
                            ? "bg-black text-white"
                            : "text-black/75 hover:bg-black/5 hover:text-black"
                          }`}
                      >
                        {jar}
                      </button>
                    ))}
                  </div>
                )}
              </div>

              {/* Min & Max Amount Fields */}
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <label className="block text-[11px] font-bold uppercase tracking-wider text-black/40 mb-1.5">
                    Min Amount (₹)
                  </label>
                  <input
                    type="number"
                    value={minAmount}
                    onChange={(e) => onMinAmountChange(e.target.value)}
                    placeholder="0"
                    className="h-11 w-full rounded-xl border border-black/10 bg-white px-3 text-xs font-bold text-black outline-none transition focus:border-black placeholder:text-black/30"
                  />
                </div>
                <div>
                  <label className="block text-[11px] font-bold uppercase tracking-wider text-black/40 mb-1.5">
                    Max Amount (₹)
                  </label>
                  <input
                    type="number"
                    value={maxAmount}
                    onChange={(e) => onMaxAmountChange(e.target.value)}
                    placeholder="Any"
                    className="h-11 w-full rounded-xl border border-black/10 bg-white px-3 text-xs font-bold text-black outline-none transition focus:border-black placeholder:text-black/30"
                  />
                </div>
              </div>

              {/* Panel Actions */}
              <div className="pt-2 flex justify-between border-t border-black/5">
                <button
                  type="button"
                  onClick={() => {
                    onCategoryFilterChange("all");
                    onJarFilterChange("all");
                    onMinAmountChange("");
                    onMaxAmountChange("");
                  }}
                  className="text-xs font-bold text-black/50 hover:text-black"
                >
                  Reset
                </button>
                <button
                  type="button"
                  onClick={() => setFilterPanelOpen(false)}
                  className="rounded-xl bg-black px-4 py-1.5 text-xs font-bold text-white"
                >
                  Apply
                </button>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Date Range Selector Bar (Using imported ui component) */}
      <DateRangePicker value={dateRange} onChange={onDateRangeChange} />
    </section>
  );
}

// Re-export utility functions so AccountStatement can keep importing them from here if needed
export { isInDateRange, toDateKey };
export type { DateRange };
