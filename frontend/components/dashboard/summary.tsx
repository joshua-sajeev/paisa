import type { DashboardSummary } from "@/lib/types/dashboard";
import { BalanceCard } from "@/components/dashboard/balance-card";
import { MonthlyPulse } from "@/components/dashboard/monthly-pulse";

type SummaryProps = {
  data: DashboardSummary;
};

export function Summary({ data }: SummaryProps) {
  return (
    <div className="px-5 pt-6">
      <BalanceCard data={data} />
      <MonthlyPulse data={data} />
    </div>
  );
}
