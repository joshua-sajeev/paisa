"use client";
import type { StatementTransaction } from "./AccountStatement";
import { useEffect, useMemo, useRef, useState } from "react";
import {
  Search,
  CalendarMonth,
  ExpandCircleDown,
} from "@material-symbols-svg/react/w400";

type Filter = "all" | "income" | "expense" | "transfer";

export type DateRange = { start: string; end: string } | null;

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

const WEEKDAYS = ["S", "M", "T", "W", "T", "F", "S"];
const MONTHS = [
  "January", "February", "March", "April", "May", "June",
  "July", "August", "September", "October", "November", "December",
];

function toKey(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
}

function fromKey(key: string): Date {
  const [y, m, d] = key.split("-").map(Number);
  return new Date(y, m - 1, d);
}

function formatShort(key: string): string {
  const d = fromKey(key);
  return `${d.getDate()} ${MONTHS[d.getMonth()].slice(0, 3)}`;
}

function formatLabel(range: DateRange): string {
  if (!range) return "All dates";
  const startD = fromKey(range.start);
  const endD = fromKey(range.end);
  if (range.start === range.end) {
    return `${formatShort(range.start)} ${startD.getFullYear()}`;
  }
  if (startD.getFullYear() === endD.getFullYear()) {
    return `${formatShort(range.start)} – ${formatShort(range.end)} ${startD.getFullYear()}`;
  }
  return `${formatShort(range.start)} ${startD.getFullYear()} – ${formatShort(range.end)} ${endD.getFullYear()}`;
}

export function toDateKey(occurredAt: string): string {
  if (!occurredAt) return "";
  if (!occurredAt.includes("T") && /^\d{4}-\d{2}-\d{2}$/.test(occurredAt)) {
    return occurredAt;
  }
  const date = new Date(occurredAt);
  if (isNaN(date.getTime())) return "";
  return toKey(date);
}

export function isInDateRange(occurredAt: string, range: DateRange): boolean {
  if (!range) return true;
  const key = toDateKey(occurredAt);
  if (!key) return false;
  return key >= range.start && key <= range.end;
}

function DateRangeCalendar({
  value,
  onChange,
  onClose,
}: {
  value: DateRange;
  onChange: (value: DateRange) => void;
  onClose: () => void;
}) {
  const initial = value ? fromKey(value.start) : new Date();
  const [viewYear, setViewYear] = useState(initial.getFullYear());
  const [viewMonth, setViewMonth] = useState(initial.getMonth());
  const [pickingEnd, setPickingEnd] = useState(false);
  const [hover, setHover] = useState<string | null>(null);

  const todayKey = toKey(new Date());

  const cells = useMemo(() => {
    const firstWeekday = new Date(viewYear, viewMonth, 1).getDay();
    const daysInMonth = new Date(viewYear, viewMonth + 1, 0).getDate();
    const list: (string | null)[] = [];
    for (let i = 0; i < firstWeekday; i++) list.push(null);
    for (let day = 1; day <= daysInMonth; day++) {
      list.push(toKey(new Date(viewYear, viewMonth, day)));
    }
    return list;
  }, [viewYear, viewMonth]);

  function shiftMonth(delta: number) {
    const next = new Date(viewYear, viewMonth + delta, 1);
    setViewYear(next.getFullYear());
    setViewMonth(next.getMonth());
  }

  function handlePick(key: string) {
    if (!pickingEnd || !value) {
      onChange({ start: key, end: key });
      setPickingEnd(true);
      return;
    }
    const start = key < value.start ? key : value.start;
    const end = key < value.start ? value.start : key;
    onChange({ start, end });
    setPickingEnd(false);
    setHover(null);
    onClose();
  }

  const previewEnd = pickingEnd && hover ? hover : null;
  const rangeStart = value
    ? previewEnd && previewEnd < value.start ? previewEnd : value.start
    : null;
  const rangeEnd = value
    ? previewEnd && previewEnd > value.start
      ? previewEnd
      : previewEnd && previewEnd < value.start
        ? value.start
        : value.end
    : null;

  return (
    <div className="absolute left-0 right-0 z-20 mt-2 rounded-2xl border border-black/10 bg-white p-4 shadow-lg">
      <div className="mb-3 flex items-center justify-between">
        <button
          type="button"
          onClick={() => shiftMonth(-1)}
          aria-label="Previous month"
          className="flex h-8 w-8 items-center justify-center rounded-full text-lg text-black/60 hover:bg-black/5"
        >
          ‹
        </button>
        <p className="text-sm font-bold">
          {MONTHS[viewMonth]} {viewYear}
        </p>
        <button
          type="button"
          onClick={() => shiftMonth(1)}
          aria-label="Next month"
          className="flex h-8 w-8 items-center justify-center rounded-full text-lg text-black/60 hover:bg-black/5"
        >
          ›
        </button>
      </div>

      <div className="grid grid-cols-7 gap-y-1 text-center">
        {WEEKDAYS.map((day, i) => (
          <span key={`${day}-${i}`} className="pb-1 text-[11px] font-semibold text-black/35">
            {day}
          </span>
        ))}

        {cells.map((key, index) => {
          if (!key) return <span key={`empty-${index}`} />;

          const isStart = key === rangeStart;
          const isEnd = key === rangeEnd;
          const inRange = rangeStart && rangeEnd && key > rangeStart && key < rangeEnd;
          const isToday = key === todayKey;
          const selected = isStart || isEnd;

          return (
            <div
              key={key}
              className={`flex justify-center ${inRange
                ? "bg-black/5"
                : isStart && rangeEnd && rangeStart !== rangeEnd
                  ? "rounded-l-full bg-black/5"
                  : isEnd && rangeStart && rangeStart !== rangeEnd
                    ? "rounded-r-full bg-black/5"
                    : ""
                }`}
            >
              <button
                type="button"
                onClick={() => handlePick(key)}
                onMouseEnter={() => pickingEnd && setHover(key)}
                className={`h-9 w-9 rounded-full text-xs font-semibold transition ${selected
                  ? "bg-black text-white"
                  : isToday
                    ? "border border-black/20 text-black"
                    : "text-black/75 hover:bg-black/5"
                  }`}
              >
                {Number(key.slice(8))}
              </button>
            </div>
          );
        })}
      </div>

      <div className="mt-3 flex items-center justify-between border-t border-black/5 pt-3">
        <button
          type="button"
          onClick={() => {
            onChange(null);
            setPickingEnd(false);
            onClose();
          }}
          className="text-xs font-bold text-black/50 hover:text-black"
        >
          Clear
        </button>
        <p className="text-[11px] text-black/40">
          {pickingEnd ? "Tap another date for a range" : "Tap a date, or two for a range"}
        </p>
        <button
          type="button"
          onClick={onClose}
          className="rounded-full bg-black px-4 py-1.5 text-xs font-bold text-white"
        >
          Done
        </button>
      </div>
    </div>
  );
}

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
  const [calendarOpen, setCalendarOpen] = useState(false);
  const [filterPanelOpen, setFilterPanelOpen] = useState(false);

  // Custom dropdown states
  const [categoryOpen, setCategoryOpen] = useState(false);
  const [jarOpen, setJarOpen] = useState(false);

  const wrapperRef = useRef<HTMLDivElement>(null);
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
      if (wrapperRef.current && !wrapperRef.current.contains(event.target as Node)) {
        setCalendarOpen(false);
      }
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

      {/* Date Range Selector Bar */}
      <div ref={wrapperRef} className="relative">
        <button
          type="button"
          onClick={() => setCalendarOpen((prev) => !prev)}
          aria-expanded={calendarOpen}
          className="flex h-11 w-full items-center justify-between rounded-2xl border border-black/10 bg-white px-4 text-sm font-semibold hover:border-black/30 transition"
        >
          <span className="flex items-center gap-2">
            <CalendarMonth className="h-5 w-5 text-black/50" />
            {formatLabel(dateRange)}
          </span>
          <ExpandCircleDown
            width={17}
            height={17}
            className={`shrink-0 text-secondary transition-transform duration-200 ${calendarOpen ? "" : "-rotate-90"
              }`}
          />

        </button>

        {calendarOpen && (
          <DateRangeCalendar
            value={dateRange}
            onChange={onDateRangeChange}
            onClose={() => setCalendarOpen(false)}
          />
        )}
      </div>
    </section>
  );
}
