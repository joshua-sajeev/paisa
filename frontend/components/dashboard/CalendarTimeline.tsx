'use client';

import { useState, useRef, useEffect } from 'react';
import CalendarExpanded from './CalendarExpanded';

export default function CalendarTimeline() {
  const [selectedDate, setSelectedDate] = useState(new Date());
  const [isExpanded, setIsExpanded] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const today = new Date();

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsExpanded(false);
      }
    };

    if (isExpanded) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isExpanded]);

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
    <div className="w-full relative" ref={containerRef}>
      {/* Compact Calendar */}
      <div className="w-full bg-white border border-gray-200/80 rounded-[22px] p-3 shadow-sm transition-all hover:shadow-md">
        <div className="flex items-center justify-between">
          {/* Today */}
          <button
            type="button"
            onClick={goToToday}
            className="flex items-center gap-2.5 md:gap-3 cursor-pointer group"
          >
            <div className="w-9 h-9 md:w-10 md:h-10 rounded-xl bg-[#e3eeff] flex items-center justify-center text-[#005bb2] transition-colors group-hover:bg-[#d0e4ff]">
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
          <div className="flex items-center gap-1 md:gap-2 bg-gray-50/80 border border-gray-100 p-1 rounded-xl">
            {/* Previous Day */}
            <button
              type="button"
              aria-label="Previous day"
              onClick={goToPreviousDay}
              className="w-7 h-7 flex items-center justify-center rounded-lg text-gray-500 hover:text-[#005bb2] hover:bg-white shadow-xs cursor-pointer active:scale-95 transition-all"
            >
              <svg
                className="w-4 h-4"
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
            <button
              type="button"
              onClick={() => setIsExpanded((current) => !current)}
              className="px-2.5 py-1 text-center text-[13px] md:text-[14px] font-bold text-[#005bb2] cursor-pointer hover:opacity-85 transition-opacity"
              aria-label="Open calendar"
            >
              {formatDate(selectedDate)}
            </button>

            {/* Next Day */}
            <button
              type="button"
              aria-label="Next day"
              onClick={goToNextDay}
              className="w-7 h-7 flex items-center justify-center rounded-lg text-gray-500 hover:text-[#005bb2] hover:bg-white shadow-xs cursor-pointer active:scale-95 transition-all"
            >
              <svg
                className="w-4 h-4"
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

      {/* Expanded Calendar Overlay */}
      {isExpanded && (
        <div className="absolute top-0 left-0 w-full z-30 shadow-xl">
          <CalendarExpanded
            selectedDate={selectedDate}
            onDateChange={setSelectedDate}
            onCollapse={() => setIsExpanded(false)}
          />
        </div>
      )}
    </div>
  );
}
