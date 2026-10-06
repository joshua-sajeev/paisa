"use client";

import { Search, Tune } from "@material-symbols-svg/react/w400";

import DateRangePicker, {
  type DateRange,
} from "@/components/ui/DateRangePicker";
import {
  CATEGORIES,
  type Account,
  type Jar,
  type TransactionCategory,
  type TransactionType,
} from "./transaction-utils";

type TransactionsFiltersProps = {
  search: string;
  type: TransactionType | "all";
  account: string;
  jar: string;
  category: TransactionCategory | "all";
  dateRange: DateRange;
  accounts: Account[];
  jars: Jar[];
  filterOpen: boolean;
  loadingFilters: boolean;
  hasActiveFilters: boolean;
  onSearchChange: (value: string) => void;
  onTypeChange: (value: TransactionType | "all") => void;
  onAccountChange: (value: string) => void;
  onJarChange: (value: string) => void;
  onCategoryChange: (value: TransactionCategory | "all") => void;
  onDateRangeChange: (value: DateRange) => void;
  onFilterToggle: () => void;
  onReset: () => void;
};

const types: { value: TransactionType | "all"; label: string }[] = [
  { value: "all", label: "All" },
  { value: "income", label: "Income" },
  { value: "expense", label: "Expense" },
  { value: "transfer", label: "Transfers" },
];

const filterColors = ["#6366F1", "#07B682", "#F59E0B", "#EC4899", "#8B5CF6"];

const typeColors: Record<TransactionType | "all", string> = {
  all: filterColors[0],
  income: filterColors[1],
  expense: filterColors[3],
  transfer: filterColors[4],
};

const selectClassName =
  "w-full min-w-0 rounded-2xl border border-black/10 bg-white px-4 py-3 text-sm text-[#0b1c30] outline-none transition focus:border-[#004ac6] focus:ring-2 focus:ring-[#004ac6]/10";

export default function TransactionsFilters({
  search,
  type,
  account,
  jar,
  category,
  dateRange,
  accounts,
  jars,
  filterOpen,
  loadingFilters,
  hasActiveFilters,
  onSearchChange,
  onTypeChange,
  onAccountChange,
  onJarChange,
  onCategoryChange,
  onDateRangeChange,
  onFilterToggle,
  onReset,
}: TransactionsFiltersProps) {
  return (
    <section className="space-y-4">
      {/* Search and filters */}
      <div className="flex items-center gap-2">
        <label className="flex h-12 min-w-0 flex-1 items-center gap-2.5 rounded-2xl border border-black/10 bg-white px-4 shadow-sm transition focus-within:border-[#6366F1] focus-within:ring-2 focus-within:ring-[#6366F1]/15">
          <Search className="h-5 w-5 shrink-0 text-black/40" />
          <input
            type="search"
            value={search}
            onChange={(event) => onSearchChange(event.target.value)}
            placeholder="Search transactions..."
            aria-label="Search transactions"
            className="w-full bg-transparent text-sm font-medium text-[#0b1c30] outline-none placeholder:text-black/35"
          />
        </label>
        <button
          type="button"
          onClick={onFilterToggle}
          aria-expanded={filterOpen}
          aria-label={filterOpen ? "Hide filters" : "Show filters"}
          className={`relative inline-flex h-12 shrink-0 items-center gap-2 rounded-2xl border px-4 text-sm font-semibold transition shadow-sm ${
            filterOpen || hasActiveFilters
              ? "border-[#6366F1] bg-[#6366F1] text-white shadow-[#6366F1]/20"
              : "border-black/10 bg-white text-[#0b1c30] hover:bg-black/[0.02]"
          }`}
        >
          <Tune className="h-5 w-5" />
          Filters
          {hasActiveFilters && <span className="h-2 w-2 rounded-full bg-white ring-2 ring-white/30" />}
        </button>
      </div>

      {/* Date Range Picker */}
      <DateRangePicker value={dateRange} onChange={onDateRangeChange} />

      {/* Quick Filters Group */}
      <div className="space-y-4 rounded-3xl border border-slate-200/80 bg-white p-4 shadow-sm">
        {/* Transaction Type Pills */}
        <div className="space-y-2">
          <div className="flex items-center justify-between px-0.5">
            <span className="text-xs font-bold text-slate-700">Type</span>
            <span className="text-[10px] font-medium uppercase tracking-wider text-slate-400">Choose one</span>
          </div>
          <div className="grid grid-cols-4 gap-2">
            {types.map((item) => (
              <button
                key={item.value}
                type="button"
                onClick={() => onTypeChange(item.value)}
                aria-pressed={type === item.value}
                style={type === item.value ? { backgroundColor: typeColors[item.value] } : undefined}
                className={`min-w-0 rounded-xl px-2 py-2.5 text-[11px] font-bold transition sm:px-3 sm:text-xs ${
                  type === item.value
                    ? "text-white shadow-sm"
                    : "border border-slate-200 bg-slate-50 text-slate-600 hover:border-slate-300 hover:bg-white"
                }`}
              >
                {item.label}
              </button>
            ))}
          </div>
        </div>

        {/* Account Pills */}
        <div className="space-y-2 border-t border-slate-100 pt-3">
          <div className="flex items-center justify-between px-0.5">
            <span className="text-xs font-bold text-slate-700">Account</span>
            <span className="text-[10px] font-medium text-slate-400">Bank or wallet</span>
          </div>
          <div className="flex flex-wrap gap-2">
            {[{ id: "all", name: "All accounts" }, ...accounts].map((item) => (
              <button
                key={item.id}
                type="button"
                disabled={loadingFilters}
                aria-pressed={account === item.id}
                onClick={() => onAccountChange(item.id)}
                style={account === item.id ? { backgroundColor: filterColors[0] } : undefined}
                className={`max-w-full rounded-xl px-3.5 py-2 text-xs font-semibold transition ${
                  account === item.id
                    ? "text-white shadow-sm"
                    : "border border-slate-200 bg-white text-slate-600 hover:border-[#6366F1]/40 hover:bg-[#6366F1]/5"
                }`}
              >
                {item.name}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Advanced Filter Drawer (Jars & Category) */}
      {filterOpen && (
        <div className="space-y-4 rounded-3xl border border-[#8B5CF6]/20 bg-gradient-to-br from-[#8B5CF6]/[0.06] via-white to-[#6366F1]/[0.04] p-4 shadow-sm animate-in fade-in slide-in-from-top-2 duration-200">
          <div className="flex items-center justify-between border-b border-slate-200/70 pb-3">
            <div>
              <h3 className="text-sm font-bold text-slate-800">More filters</h3>
              <p className="mt-0.5 text-xs text-slate-500">Narrow transactions by jar or category</p>
            </div>
            {hasActiveFilters && (
              <button
                type="button"
                onClick={onReset}
                className="rounded-lg px-2.5 py-1.5 text-xs font-bold text-[#6366F1] transition hover:bg-[#6366F1]/10"
              >
                Clear all
              </button>
            )}
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
          <div className="min-w-0 space-y-2" aria-label="Filter by jar">
            <span className="px-0.5 text-xs font-bold text-slate-700">Jar</span>
            <div className="flex flex-wrap gap-2">
              {[{ id: "all", name: "All jars" }, ...jars].map((item, index) => (
                <button
                  key={item.id}
                  type="button"
                  disabled={loadingFilters}
                  aria-pressed={jar === item.id}
                  onClick={() => onJarChange(item.id)}
                  style={jar === item.id ? { backgroundColor: filterColors[index % filterColors.length] } : undefined}
                  className={`max-w-full rounded-xl px-3 py-2 text-xs font-semibold transition ${
                    jar === item.id
                      ? "text-white shadow-sm"
                      : "border border-slate-200 bg-white text-slate-600 hover:border-[#8B5CF6]/40 hover:bg-white"
                  }`}
                >
                  {item.name}
                </button>
              ))}
            </div>
          </div>

          <div className="min-w-0 space-y-2">
            <span className="px-0.5 text-xs font-bold text-slate-700">Category</span>
            <select
              value={category}
              onChange={(event) =>
                onCategoryChange(event.target.value as TransactionCategory | "all")
              }
              aria-label="Filter by category"
              className={`${selectClassName} border-slate-200 bg-white shadow-sm focus:border-[#8B5CF6] focus:ring-[#8B5CF6]/15`}
            >
              {CATEGORIES.map((item) => (
                <option key={item.value} value={item.value}>
                  {item.label}
                </option>
              ))}
            </select>
          </div>
          </div>
        </div>
      )}
    </section>
  );
}
