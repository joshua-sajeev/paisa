"use client";

import { useMemo, useState } from "react";
import {
  CalendarMonth,
  ExpandCircleDown,
} from "@material-symbols-svg/react/w400";

export type DateRange = {
  start: string;
  end: string;
} | null;

const WEEKDAYS = ["S", "M", "T", "W", "T", "F", "S"];

const MONTHS = [
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December",
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
  const date = fromKey(key);

  return `${date.getDate()} ${MONTHS[date.getMonth()].slice(0, 3)}`;
}

function formatLabel(range: DateRange): string {
  if (!range) return "All dates";

  const startDate = fromKey(range.start);
  const endDate = fromKey(range.end);

  if (range.start === range.end) {
    return `${formatShort(range.start)} ${startDate.getFullYear()}`;
  }

  if (startDate.getFullYear() === endDate.getFullYear()) {
    return `${formatShort(range.start)} – ${formatShort(range.end)} ${startDate.getFullYear()}`;
  }

  return `${formatShort(range.start)} ${startDate.getFullYear()} – ${formatShort(range.end)} ${endDate.getFullYear()}`;
}

export function toDateKey(occurredAt: string): string {
  if (!occurredAt) return "";

  if (
    !occurredAt.includes("T") &&
    /^\d{4}-\d{2}-\d{2}$/.test(occurredAt)
  ) {
    return occurredAt;
  }

  const date = new Date(occurredAt);

  if (isNaN(date.getTime())) return "";

  return toKey(date);
}

export function isInDateRange(
  occurredAt: string,
  range: DateRange,
): boolean {
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

    const daysInMonth = new Date(
      viewYear,
      viewMonth + 1,
      0,
    ).getDate();

    const list: (string | null)[] = [];

    for (let i = 0; i < firstWeekday; i++) {
      list.push(null);
    }

    for (let day = 1; day <= daysInMonth; day++) {
      list.push(
        toKey(new Date(viewYear, viewMonth, day)),
      );
    }

    return list;
  }, [viewYear, viewMonth]);

  function shiftMonth(delta: number) {
    const next = new Date(
      viewYear,
      viewMonth + delta,
      1,
    );

    setViewYear(next.getFullYear());
    setViewMonth(next.getMonth());
  }

  function handlePick(key: string) {
    if (!pickingEnd || !value) {
      onChange({
        start: key,
        end: key,
      });

      setPickingEnd(true);
      return;
    }

    const start =
      key < value.start ? key : value.start;

    const end =
      key < value.start ? value.start : key;

    onChange({
      start,
      end,
    });

    setPickingEnd(false);
    setHover(null);
    onClose();
  }

  const previewEnd =
    pickingEnd && hover ? hover : null;

  const rangeStart = value
    ? previewEnd && previewEnd < value.start
      ? previewEnd
      : value.start
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
        {WEEKDAYS.map((day, index) => (
          <span
            key={`${day}-${index}`}
            className="pb-1 text-[11px] font-semibold text-black/35"
          >
            {day}
          </span>
        ))}

        {cells.map((key, index) => {
          if (!key) {
            return (
              <span key={`empty-${index}`} />
            );
          }

          const isStart = key === rangeStart;
          const isEnd = key === rangeEnd;

          const inRange =
            rangeStart &&
            rangeEnd &&
            key > rangeStart &&
            key < rangeEnd;

          const isToday = key === todayKey;

          const selected = isStart || isEnd;

          return (
            <div
              key={key}
              className={`flex justify-center ${
                inRange
                  ? "bg-black/5"
                  : isStart &&
                      rangeEnd &&
                      rangeStart !== rangeEnd
                    ? "rounded-l-full bg-black/5"
                    : isEnd &&
                        rangeStart &&
                        rangeStart !== rangeEnd
                      ? "rounded-r-full bg-black/5"
                      : ""
              }`}
            >
              <button
                type="button"
                onClick={() => handlePick(key)}
                onMouseEnter={() =>
                  pickingEnd && setHover(key)
                }
                className={`h-9 w-9 rounded-full text-xs font-semibold transition ${
                  selected
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
            setHover(null);
            onClose();
          }}
          className="text-xs font-bold text-black/50 hover:text-black"
        >
          Clear
        </button>

        <p className="text-[11px] text-black/40">
          {pickingEnd
            ? "Tap another date for a range"
            : "Tap a date, or two for a range"}
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

type Props = {
  value: DateRange;
  onChange: (value: DateRange) => void;
};

export default function DateRangePicker({
  value,
  onChange,
}: Props) {
  const [calendarOpen, setCalendarOpen] =
    useState(false);

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() =>
          setCalendarOpen((prev) => !prev)
        }
        aria-expanded={calendarOpen}
        className="flex h-11 w-full items-center justify-between rounded-2xl border border-black/10 bg-white px-4 text-sm font-semibold transition hover:border-black/30"
      >
        <span className="flex items-center gap-2">
          <CalendarMonth className="h-5 w-5 text-black/50" />

          {formatLabel(value)}
        </span>

        <ExpandCircleDown
          width={17}
          height={17}
          className={`shrink-0 text-secondary transition-transform duration-200 ${
            calendarOpen ? "" : "-rotate-90"
          }`}
        />
      </button>

      {calendarOpen && (
        <DateRangeCalendar
          value={value}
          onChange={onChange}
          onClose={() =>
            setCalendarOpen(false)
          }
        />
      )}
    </div>
  );
}
