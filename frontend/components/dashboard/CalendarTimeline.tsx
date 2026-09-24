'use client';

import { useState } from 'react';

export default function CalendarTimeline() {
  const [selectedDate, setSelectedDate] = useState(new Date());

  const today = new Date();

  const formatDate = (date: Date) => {
    return date.toLocaleDateString('en-GB', {
      day: 'numeric',
      month: 'short',
    });
  };

  const goToPreviousDay = () => {
    setSelectedDate((current) => {
      const previous = new Date(current);
      previous.setDate(previous.getDate() - 1);
      return previous;
    });
  };

  const goToNextDay = () => {
    setSelectedDate((current) => {
      const next = new Date(current);
      next.setDate(next.getDate() + 1);
      return next;
    });
  };

  const goToToday = () => {
    setSelectedDate(new Date());
  };

  return (
    <div className="w-full bg-white border border-gray-200/80 rounded-[22px] p-3 shadow-sm transition-all hover:shadow-md">
      <div className="flex items-center justify-between">

        {/* Today */}
        <button
          type="button"
          onClick={goToToday}
          className="flex items-center gap-2.5 md:gap-3 cursor-pointer"
        >
          <div className="w-9 h-9 md:w-10 md:h-10 rounded-2xl bg-[#e3eeff] flex items-center justify-center text-[#005bb2]">
            <svg
              className="w-4 h-4 md:w-5 md:h-5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
              />
            </svg>
          </div>

          <span className="text-[14px] md:text-[17px] font-bold text-[#1c1b1f] tracking-tight">
            Today, {formatDate(today)}
          </span>
        </button>

        {/* Date Navigation */}
        <div className="flex items-center gap-2 md:gap-3">

          {/* Previous Day */}
          <button
            type="button"
            aria-label="Previous day"
            onClick={goToPreviousDay}
            className="p-1 text-gray-400 hover:text-[#005bb2] cursor-pointer active:scale-90 transition-colors"
          >
            <svg
              className="w-4 h-4 md:w-5 md:h-5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2.5}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M15 19l-7-7 7-7"
              />
            </svg>
          </button>

          {/* Selected Date */}
          <span className="min-w-[52px] text-center text-[13px] md:text-[14px] font-bold text-[#005bb2]">
            {formatDate(selectedDate)}
          </span>

          {/* Next Day */}
          <button
            type="button"
            aria-label="Next day"
            onClick={goToNextDay}
            className="p-1 text-gray-400 hover:text-[#005bb2] cursor-pointer active:scale-90 transition-colors"
          >
            <svg
              className="w-4 h-4 md:w-5 md:h-5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2.5}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M9 5l7 7-7 7"
              />
            </svg>
          </button>

        </div>
      </div>
    </div>
  );
}
