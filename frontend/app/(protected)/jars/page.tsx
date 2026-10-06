
"use client";

import { useState } from "react";

import AllocationHistory from "@/components/jars/AllocationHistory";
import DateRangePicker, {
  type DateRange,
} from "@/components/ui/DateRangePicker";

export default function JarsPage() {
  const [dateRange, setDateRange] = useState<DateRange>(null);

  return (
    <main className="min-h-screen bg-surface pb-24">
      <div className="mx-auto flex w-full max-w-md flex-col gap-3 px-4">
        <DateRangePicker
          value={dateRange}
          onChange={setDateRange}
        />

        <AllocationHistory dateRange={dateRange} />
      </div>
    </main>
  );
}
