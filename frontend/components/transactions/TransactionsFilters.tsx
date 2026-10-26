"use client";

import { Search, Tune } from "@material-symbols-svg/react/w400";
import { useCallback, useEffect, useRef, useState } from "react";

import DateRangePicker, { type DateRange } from "@/components/ui/DateRangePicker";
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
  minAmount: string;
  maxAmount: string;
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
  onMinAmountChange: (value: string) => void;
  onMaxAmountChange: (value: string) => void;
  onFilterToggle: () => void;
  onReset: () => void;
};

const types: { value: TransactionType | "all"; label: string }[] = [
  { value: "all", label: "All" },
  { value: "income", label: "Income" },
  { value: "expense", label: "Expense" },
  { value: "transfer", label: "Transfers" },
];

const selectClassName =
  "w-full min-w-0 rounded-xl border border-black/10 bg-white px-3 py-2.5 text-xs font-bold text-black outline-none transition focus:border-black";

export default function TransactionsFilters({
  search,
  type,
  account,
  jar,
  category,
  dateRange,
  minAmount,
  maxAmount,
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
  onMinAmountChange,
  onMaxAmountChange,
  onFilterToggle,
  onReset,
}: TransactionsFiltersProps) {
  const [draftCategory, setDraftCategory] = useState(category);
  const [draftAccount, setDraftAccount] = useState(account);
  const [draftJar, setDraftJar] = useState(jar);
  const [draftDateRange, setDraftDateRange] = useState(dateRange);
  const [draftMinAmount, setDraftMinAmount] = useState(minAmount);
  const [draftMaxAmount, setDraftMaxAmount] = useState(maxAmount);
  const [categoryOpen, setCategoryOpen] = useState(false);
  const [accountOpen, setAccountOpen] = useState(false);
  const [jarOpen, setJarOpen] = useState(false);
  const filterPanelRef = useRef<HTMLDivElement>(null);

  const toggleFilters = useCallback(() => {
    if (!filterOpen) {
      setDraftCategory(category);
      setDraftAccount(account);
      setDraftJar(jar);
      setDraftDateRange(dateRange);
      setDraftMinAmount(minAmount);
      setDraftMaxAmount(maxAmount);
    }
    onFilterToggle();
  }, [filterOpen, category, account, jar, dateRange, minAmount, maxAmount, onFilterToggle]);

  useEffect(() => {
    function handleOutside(event: MouseEvent | TouchEvent) {
      if (filterPanelRef.current && !filterPanelRef.current.contains(event.target as Node)) {
        if (filterOpen) toggleFilters();
        setCategoryOpen(false);
        setAccountOpen(false);
        setJarOpen(false);
      }
    }
    document.addEventListener("mousedown", handleOutside);
    document.addEventListener("touchstart", handleOutside);
    return () => {
      document.removeEventListener("mousedown", handleOutside);
      document.removeEventListener("touchstart", handleOutside);
    };
  }, [filterOpen, toggleFilters]);

  function applyAdvancedFilters() {
    onCategoryChange(draftCategory);
    onAccountChange(draftAccount);
    onJarChange(draftJar);
    onDateRangeChange(draftDateRange);
    onMinAmountChange(draftMinAmount);
    onMaxAmountChange(draftMaxAmount);
    onFilterToggle();
  }

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
        <div className="relative" ref={filterPanelRef}>
        <button
          type="button"
          onClick={toggleFilters}
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
        {filterOpen && (
          <div className="absolute right-0 z-30 mt-2 max-h-[min(80vh,36rem)] w-80 max-w-[calc(100vw-2rem)] space-y-4 overflow-y-auto rounded-2xl border border-black/10 bg-white p-4 shadow-xl">
            <h3 className="border-b border-black/5 pb-3 text-sm font-bold text-black">Filters</h3>

            <div className="relative">
              <label className="mb-1.5 block text-[11px] font-bold uppercase tracking-wider text-black/40">Category</label>
              <button type="button" onClick={() => { setCategoryOpen((open) => !open); setAccountOpen(false); setJarOpen(false); }} className="flex h-11 w-full items-center justify-between rounded-xl border border-black/10 bg-white px-3 text-xs font-bold text-black transition hover:border-black/30">
                <span className="truncate">{CATEGORIES.find((item) => item.value === draftCategory)?.label ?? "All categories"}</span>
                <span className={`text-black/40 transition ${categoryOpen ? "rotate-180" : ""}`}>⌄</span>
              </button>
              {categoryOpen && (
                <div className="absolute left-0 right-0 z-40 mt-1 max-h-48 space-y-0.5 overflow-y-auto rounded-xl border border-black/10 bg-white p-1 shadow-lg">
                  {CATEGORIES.map((item) => (
                    <button key={item.value} type="button" onClick={() => { setDraftCategory(item.value); setCategoryOpen(false); }} className={`w-full truncate rounded-lg px-3 py-2 text-left text-xs font-bold transition ${draftCategory === item.value ? "bg-black text-white" : "text-black/75 hover:bg-black/5 hover:text-black"}`}>{item.label}</button>
                  ))}
                </div>
              )}
            </div>

            <div className="relative">
              <label className="mb-1.5 block text-[11px] font-bold uppercase tracking-wider text-black/40">Account</label>
              <button type="button" disabled={loadingFilters} onClick={() => { setAccountOpen((open) => !open); setCategoryOpen(false); setJarOpen(false); }} className="flex h-11 w-full items-center justify-between rounded-xl border border-black/10 bg-white px-3 text-xs font-bold text-black transition hover:border-black/30 disabled:opacity-50">
                <span className="truncate">{draftAccount === "all" ? "All accounts" : accounts.find((item) => item.id === draftAccount)?.name ?? "All accounts"}</span>
                <span className={`text-black/40 transition ${accountOpen ? "rotate-180" : ""}`}>⌄</span>
              </button>
              {accountOpen && (
                <div className="absolute left-0 right-0 z-40 mt-1 max-h-48 space-y-0.5 overflow-y-auto rounded-xl border border-black/10 bg-white p-1 shadow-lg">
                  {[{ id: "all", name: "All accounts" }, ...accounts].map((item) => (
                    <button key={item.id} type="button" onClick={() => { setDraftAccount(item.id); setAccountOpen(false); }} className={`flex w-full items-center justify-between truncate rounded-lg px-3 py-2 text-left text-xs font-bold transition ${draftAccount === item.id ? "bg-black text-white" : "text-black/75 hover:bg-black/5 hover:text-black"}`}>
                      <span className="truncate">{item.name}</span>
                      {"is_primary" in item && item.is_primary && <span className={`ml-2 shrink-0 text-[10px] ${draftAccount === item.id ? "text-white/70" : "text-black/40"}`}>Primary</span>}
                    </button>
                  ))}
                </div>
              )}
            </div>

            <div className="relative">
              <label className="mb-1.5 block text-[11px] font-bold uppercase tracking-wider text-black/40">Jar</label>
              <button type="button" disabled={loadingFilters} onClick={() => { setJarOpen((open) => !open); setCategoryOpen(false); setAccountOpen(false); }} className="flex h-11 w-full items-center justify-between rounded-xl border border-black/10 bg-white px-3 text-xs font-bold text-black transition hover:border-black/30 disabled:opacity-50">
                <span className="truncate">{draftJar === "all" ? "All jars" : jars.find((item) => item.id === draftJar)?.name ?? "All jars"}</span>
                <span className={`text-black/40 transition ${jarOpen ? "rotate-180" : ""}`}>⌄</span>
              </button>
              {jarOpen && (
                <div className="absolute left-0 right-0 z-40 mt-1 max-h-48 space-y-0.5 overflow-y-auto rounded-xl border border-black/10 bg-white p-1 shadow-lg">
                  {[{ id: "all", name: "All jars" }, ...jars].map((item) => (
                    <button key={item.id} type="button" onClick={() => { setDraftJar(item.id); setJarOpen(false); }} className={`w-full truncate rounded-lg px-3 py-2 text-left text-xs font-bold transition ${draftJar === item.id ? "bg-black text-white" : "text-black/75 hover:bg-black/5 hover:text-black"}`}>{item.name}</button>
                  ))}
                </div>
              )}
            </div>

            <div>
              <span className="mb-1.5 block text-[11px] font-bold uppercase tracking-wider text-black/40">Date range</span>
              <DateRangePicker value={draftDateRange} onChange={setDraftDateRange} />
            </div>

            <div className="grid grid-cols-2 gap-2">
              <label>
                <span className="mb-1.5 block text-[11px] font-bold uppercase tracking-wider text-black/40">Min amount (₹)</span>
                <input type="number" min="0" inputMode="decimal" value={draftMinAmount} onChange={(event) => setDraftMinAmount(event.target.value)} placeholder="Any" className={selectClassName} />
              </label>
              <label>
                <span className="mb-1.5 block text-[11px] font-bold uppercase tracking-wider text-black/40">Max amount (₹)</span>
                <input type="number" min="0" inputMode="decimal" value={draftMaxAmount} onChange={(event) => setDraftMaxAmount(event.target.value)} placeholder="Any" className={selectClassName} />
              </label>
            </div>

            <div className="flex items-center justify-between border-t border-black/5 pt-3">
              <button type="button" onClick={onReset} className="text-xs font-bold text-black/50 hover:text-black">Reset</button>
              <button type="button" onClick={applyAdvancedFilters} className="rounded-xl bg-black px-4 py-1.5 text-xs font-bold text-white">Apply</button>
            </div>
          </div>
        )}
        </div>
      </div>

      {/* Quick Filters Group */}
      <div className="space-y-4">
        {/* Transaction Type Pills */}
        <div>
          <div className="flex gap-2 overflow-x-auto pb-1">
            {types.map((item) => (
              <button
                key={item.value}
                type="button"
                onClick={() => onTypeChange(item.value)}
                aria-pressed={type === item.value}
                className={`shrink-0 rounded-full px-4 py-2 text-xs font-bold transition ${
                  type === item.value
                    ? "bg-black text-white"
                    : "border border-black/10 bg-white text-black/60 hover:border-black/30"
                }`}
              >
                {item.label}
              </button>
            ))}
          </div>
        </div>

      </div>

    </section>
  );
}
