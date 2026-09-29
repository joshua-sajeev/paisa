'use client';

import { useState, useRef, useEffect } from 'react';
import {
  Today,
  ChevronLeft,
  ChevronRight,
} from '@material-symbols-svg/react/w400';
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
              <Today
                width="20"
                height="20"
                className="md:w-6 md:h-6"
              />
            </div>

            <span className="text-[14px] md:text-[17px] font-bold text-[#1c1b1f] tracking-tight">
              Today, {formatDate(today)}
            </span>
          </button>

          {/* Date Navigation */}
          <div className="flex items-center gap-1 md:gap-2">
            {/* Previous Day */}
            <button
              type="button"
              aria-label="Previous day"
              onClick={goToPreviousDay}
              className="w-7 h-7 flex items-center justify-center rounded-lg text-gray-500 hover:text-[#005bb2] cursor-pointer active:scale-95 transition-all"
            >
              <ChevronLeft width="20" height="20" />
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
              className="w-7 h-7 flex items-center justify-center rounded-lg text-gray-500 hover:text-[#005bb2] cursor-pointer active:scale-95 transition-all"
            >
              <ChevronRight width="20" height="20" />
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
