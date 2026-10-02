import { ReceiptLong } from "@material-symbols-svg/react/icons/receipt-long";
import { SwapHoriz } from "@material-symbols-svg/react/icons/swap-horiz";

type Props = {
  balance: number;
  activeCount: number;
  archivedCount: number;
};

function formatBalance(value: number) {
  return new Intl.NumberFormat("en-IN", {
    style: "currency",
    currency: "INR",
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value);
}

export default function TotalBalanceCard({
  balance,
  activeCount,
  archivedCount,
}: Props) {
  const formatted = formatBalance(balance);
  const [whole, decimal] = formatted.split(".");

  return (
    <section className="mb-5 rounded-2xl bg-white p-5 shadow-sm">
      <div className="mb-2 flex items-center justify-between gap-2">
        <span className="shrink-0 text-[11px] font-bold uppercase tracking-wider text-on-surface-variant">
          Total Balance
        </span>

        <div className="flex shrink-0 gap-1.5">
          <span className="rounded-full bg-[#e6f9f3] px-2 py-0.5 text-[11px] font-bold uppercase tracking-wide text-[#07b682]">
            {activeCount} Active
          </span>

          <span className="rounded-full bg-[#f1f3f5] px-2 py-0.5 text-[11px] font-bold uppercase text-secondary">
            {archivedCount} Archived
          </span>
        </div>
      </div>

      <div className="my-2">
        <div className="font-extrabold tracking-tight tabular-nums text-on-surface">
          <span className="text-[34px] leading-none">
            {whole}
          </span>

          <span className="ml-0.5 text-xl font-bold text-on-surface-variant">
            .{decimal}
          </span>
        </div>
      </div>

      <div className="mt-4">
        <button
          type="button"
          className="flex w-full items-center justify-center gap-1.5 rounded-xl bg-[#eaf3ff] px-2 py-2.5 text-xs font-semibold text-[#2563eb] transition hover:bg-[#dbeafe] active:scale-95"
        >
          <SwapHoriz width={16} height={16} />
          <span>Transfer</span>
        </button>
      </div>
    </section>
  );
}
