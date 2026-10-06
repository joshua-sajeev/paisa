"use client";

import { useState } from "react";

import JarsList from "@/components/jars/JarsList";
import AllocationHistory from "@/components/jars/AllocationHistory";
import DateRangePicker, {
  type DateRange,
} from "@/components/ui/DateRangePicker";

export default function JarsPage() {
  const [dateRange, setDateRange] = useState<DateRange>(null);

  return (
    <main className="min-h-full bg-surface p-3 max-w-md mx-auto flex flex-col gap-3">
      <div className="flex flex-col gap-3 w-full pb-4">
        <DateRangePicker
          value={dateRange}
          onChange={setDateRange}
        />

        <JarsList dateRange={dateRange} />

        <AllocationHistory dateRange={dateRange} />
      </div>
    </main>
  );
}
