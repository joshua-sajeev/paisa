import CalendarTimeline from "@/components/dashboard/CalendarTimeline";

export default function DashboardPage() {
  return (
    <main className="min-h-full p-3 max-w-5xl mx-auto flex flex-col gap-3">
      <div>
        <CalendarTimeline />
      </div>
    </main>
  );
}
