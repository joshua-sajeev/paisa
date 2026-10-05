import { Allocation, AllocationType } from "@/lib/jars";
import { formatDate, formatMoney } from "@/lib/utils";

type Props = {
  allocation: Allocation;
  color: string;
};

const allocationTypeLabel: Record<AllocationType, string> = {
  percentage: "Percentage",
  fixed: "Fixed",
  remainder: "Remainder",
};

export function AllocationRow({ allocation, color }: Props) {
  const date = formatDate(allocation.occurred_at);

  return (
    <div className="flex items-center justify-between gap-3 p-3 transition-colors hover:bg-slate-50">
      {/* Date + transaction + jar */}
      <div className="flex min-w-0 items-center gap-2.5">
        {/* Date */}
        <div
          className="flex h-10 w-10 shrink-0 flex-col items-center justify-center rounded-xl border"
          style={{
            color,
            backgroundColor: `${color}12`,
            borderColor: `${color}25`,
          }}
        >
          <span className="text-xs font-black leading-none">
            {date.day}
          </span>

          <span className="mt-0.5 text-[9px] font-bold uppercase leading-none">
            {date.month}
          </span>
        </div>

        {/* Transaction + Jar */}
        <div className="min-w-0">
          <div className="truncate text-xs font-bold text-slate-900">
            {allocation.transaction_name}
          </div>

          <div className="mt-0.5 flex min-w-0 items-center gap-1.5 text-[11px] font-semibold text-slate-600">
            <span
              className="h-1.5 w-1.5 shrink-0 rounded-full"
              style={{ backgroundColor: color }}
            />

            <span className="truncate">
              {allocation.jar_name}
            </span>
          </div>
        </div>
      </div>

      {/* Amount + allocation rule */}
      <div className="shrink-0 pl-2 text-right">
        <div
          className="text-sm font-black"
          style={{ color }}
        >
          +₹{formatMoney(allocation.amount)}
        </div>

        <div className="mt-0.5 text-[10px] font-bold text-slate-400">
          {allocationTypeLabel[allocation.allocation_type]}
        </div>
      </div>
    </div>
  );
}
