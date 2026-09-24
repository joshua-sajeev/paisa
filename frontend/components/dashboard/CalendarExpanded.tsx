'use client';

import { useMemo, useState } from 'react';

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
    <div className="w-full max-w-md mx-auto flex flex-col gap-3 p-4 bg-white border border-gray-200 rounded-2xl shadow-sm">
      {/* Header */}
      <div className="flex items-center justify-between pb-2 border-b border-gray-100">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-blue-600 text-white flex items-center justify-center">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="w-4 h-4"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
              <line x1="16" y1="2" x2="16" y2="6" />
              <line x1="8" y1="2" x2="8" y2="6" />
              <line x1="3" y1="10" x2="21" y2="10" />
            </svg>
          </div>
          <div>
            <h3 className="text-sm font-bold text-gray-900 leading-tight">Calendar</h3>
          </div>
        </div>
        {onCollapse && (
          <button
            type="button"
            onClick={onCollapse}
            aria-label="Collapse calendar"
            className="w-7 h-7 rounded-lg bg-gray-100 hover:bg-gray-200 flex items-center justify-center text-gray-600 transition"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="w-4 h-4"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <polyline points="18 15 12 9 6 15" />
            </svg>
          </button>
        )}
      </div>

      {/* Week Navigation */}
      <div className="flex items-center justify-between bg-gray-50 rounded-xl px-2 py-1.5 border border-gray-200/40">
        <button
          type="button"
          onClick={() => shiftWeek(-7)}
          aria-label="Previous week"
          className="flex items-center gap-1 px-2 py-1 rounded-lg text-[12px] font-bold text-blue-600 hover:bg-gray-200/60 active:scale-95 transition-all"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="w-4 h-4"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <polyline points="15 18 9 12 15 6" />
          </svg>
        </button>
        <span className="text-[14px] font-extrabold text-gray-900 tracking-tight">
          {formatWeek()}
        </span>
        <button
          type="button"
          onClick={() => shiftWeek(7)}
          aria-label="Next week"
          className="flex items-center gap-1 px-2 py-1 rounded-lg text-[12px] font-bold text-blue-600 hover:bg-gray-200/60 active:scale-95 transition-all"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="w-4 h-4"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <polyline points="9 18 15 12 9 6" />
          </svg>
        </button>
      </div>

      {/* Week Days */}
      <div className="grid grid-cols-7 gap-1 pt-1">
        {weekDays.map((date) => {
          const isToday = isSameDay(date, today);
          const isSelected = isSameDay(date, selectedDate);

          let btnStyle = "flex flex-col items-center justify-center py-2 px-1 rounded-xl transition-all cursor-pointer ";

          if (isSelected && isToday) {
            btnStyle += "bg-rose-500 shadow-sm";
          } else if (isSelected) {
            btnStyle += "bg-blue-600 text-white shadow-sm";
          } else if (isToday) {
            btnStyle += "bg-rose-500 shadow-sm";
          } else {
            btnStyle += "bg-white hover:bg-gray-100/60 border border-transparent";
          }

          return (
            <button
              key={date.toISOString()}
              type="button"
              onClick={() => onDateChange(date)}
              className={btnStyle}
            >
              <span className={`text-[14px] sm:text-[12px] font-semibold uppercase tracking-wider ${isSelected || isToday ? 'text-white font-extrabold' : 'text-gray-500'}`}>
                {date.toLocaleDateString('en-GB', { weekday: 'short' })}
              </span>
              <span className={`text-[14px] mt-1 ${isSelected || isToday ? 'text-white font-extrabold' : 'font-bold text-gray-900'}`}>
                {date.getDate()}
              </span>
            </button>
          );
        })}
      </div>

      {/* Quick Actions (Centered) */}
      <div className="flex items-center justify-center gap-1.5 pt-1 flex-wrap pb-0.5">
        <button
          type="button"
          onClick={() => onDateChange(new Date())}
          className="px-3 py-1.5 rounded-full text-[12px] font-bold text-white bg-rose-500 flex-shrink-0 transition-transform active:scale-95 shadow-sm"
        >
          Today ({today.getDate()} {today.toLocaleDateString('en-GB', { month: 'short' })})
        </button>
        <button
          type="button"
          onClick={() => onDateChange(new Date())}
          className="px-3 py-1.5 rounded-full text-[12px] font-bold bg-blue-600 text-white flex-shrink-0 transition-transform active:scale-95 shadow-sm"
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
          className="px-3 py-1.5 rounded-full text-[12px] font-semibold bg-gray-100/70 text-gray-600 hover:text-gray-900 flex-shrink-0 border border-gray-200/40 active:scale-95"
        >
          Last Week
        </button>
      </div>
    </div>
  );
}
