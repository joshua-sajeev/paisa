'use client';

import { useMemo, useState } from 'react';
import {
  CalendarToday,
  ExpandCircleUp,
  ChevronLeft,
  ChevronRight,
} from '@material-symbols-svg/react/w400';

type CalendarExpandedProps = {
  selectedDate: Date;
  onDateChange(date: Date): void;
  onCollapse?(): void;
};

export default function CalendarExpanded({
  selectedDate,
  onDateChange,
  onCollapse,
}: CalendarExpandedProps) {
  const [today] = useState(() => new Date());

  const startOfWeek = useMemo(() => {
    const date = new Date(selectedDate);
    const day = date.getDay();
    const diff = day === 0 ? -6 : 1 - day;
    date.setDate(date.getDate() + diff);
    date.setHours(0, 0, 0, 0);
    return date;
  }, [selectedDate]);

  const weekDays = useMemo(() => {
    return Array.from({ length: 7 }, (_, index) => {
      const date = new Date(startOfWeek);
      date.setDate(startOfWeek.getDate() + index);
      return date;
    });
  }, [startOfWeek]);

  const isSameDay = (a: Date, b: Date) =>
    a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate();

  const formatWeek = () => {
    const start = weekDays[0];
    const end = weekDays[6];
    const sMonth = start.toLocaleDateString('en-GB', { month: 'short' });
    const eMonth = end.toLocaleDateString('en-GB', { month: 'short' });
    const year = end.getFullYear();

    return sMonth === eMonth
      ? `Week of ${start.getDate()} – ${end.getDate()} ${eMonth}, ${year}`
      : `Week of ${start.getDate()} ${sMonth} – ${end.getDate()} ${eMonth}, ${year}`;
  };

  const shiftWeek = (days: number) => {
    const d = new Date(selectedDate);
    d.setDate(d.getDate() + days);
    onDateChange(d);
  };

  return (
    <div className="w-full flex flex-col gap-3 p-4 bg-white border border-gray-200 rounded-2xl shadow-sm">
      {/* Header */}
      <div className="flex items-center justify-between pb-2 border-b border-gray-100">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-blue-600 text-white flex items-center justify-center">
            <CalendarToday width={16} height={16} />
          </div>
          <div>
            <h3 className="text-sm font-bold text-gray-900 leading-tight">Calendar &amp; Timeline</h3>
            <p className="text-[10px] text-gray-500">Select day or navigate weekly</p>
          </div>
        </div>
        {onCollapse && (
          <button
            type="button"
            onClick={onCollapse}
            aria-label="Collapse calendar"
            className="w-7 h-7 rounded-lg bg-gray-100 hover:bg-gray-200 flex items-center justify-center text-gray-600 transition"
          >
            <ExpandCircleUp width={20} height={20} />
          </button>
        )}
      </div>

      {/* Week Navigation */}
      <div className="flex items-center justify-between bg-gray-50 rounded-xl px-2 py-1.5 border border-gray-100">
        <button
          type="button"
          onClick={() => shiftWeek(-7)}
          aria-label="Previous week"
          className="p-1 rounded-md text-blue-600 hover:bg-gray-200 transition"
        >
          <ChevronLeft width={20} height={20} />
        </button>
        <span className="text-xs font-bold text-gray-800">{formatWeek()}</span>
        <button
          type="button"
          onClick={() => shiftWeek(7)}
          aria-label="Next week"
          className="p-1 rounded-md text-blue-600 hover:bg-gray-200 transition"
        >
          <ChevronRight width={20} height={20} />
        </button>
      </div>

      {/* Week Days */}
      <div className="grid grid-cols-7 gap-1">
        {weekDays.map((date) => {
          const isToday = isSameDay(date, today);
          const isSelected = isSameDay(date, selectedDate);

          let btnStyle = "flex flex-col items-center justify-center py-2 rounded-xl transition cursor-pointer ";
          if (isSelected) btnStyle += "bg-blue-600 text-white font-bold shadow";
          else if (isToday) btnStyle += "bg-rose-500 text-white font-bold shadow";
          else btnStyle += "bg-gray-50 hover:bg-gray-100 text-gray-800";

          return (
            <button
              key={date.toISOString()}
              type="button"
              onClick={() => onDateChange(date)}
              className={btnStyle}
            >
              <span className={`text-[10px] uppercase ${isSelected || isToday ? 'text-white' : 'text-gray-400'}`}>
                {date.toLocaleDateString('en-GB', { weekday: 'short' })}
              </span>
              <span className="text-sm mt-0.5">{date.getDate()}</span>
            </button>
          );
        })}
      </div>

      {/* Quick Actions */}
      <div className="flex items-center gap-1.5 overflow-x-auto pb-1" style={{ scrollbarWidth: 'none' }}>
        <button
          type="button"
          onClick={() => onDateChange(new Date())}
          className="px-3 py-1.5 rounded-full text-xs font-bold text-white bg-rose-500 shrink-0 transition active:scale-95"
        >
          Today ({today.getDate()} {today.toLocaleDateString('en-GB', { month: 'short' })})
        </button>
        <button
          type="button"
          onClick={() => onDateChange(new Date())}
          className="px-3 py-1.5 rounded-full text-xs font-bold text-white bg-blue-600 shrink-0 transition active:scale-95"
        >
          This Week
        </button>
        <button
          type="button"
          onClick={() => {
            const d = new Date();
            d.setDate(d.getDate() - 7);
            onDateChange(d);
          }}
          className="px-3 py-1.5 rounded-full text-xs font-medium text-gray-600 bg-gray-100 hover:bg-gray-200 shrink-0 transition active:scale-95"
        >
          Last Week
        </button>
      </div>
    </div>
  );
}
