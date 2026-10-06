import { PAGE_SIZE_OPTIONS } from "./transaction-utils";

type Props = {
  page: number;
  totalPages: number;
  limit: number;
  offset: number;
  transactionCount: number;
  total: number;
  onPageChange: (page: number) => void;
  onLimitChange: (limit: number) => void;
};

export default function TransactionsPagination({
  page,
  totalPages,
  limit,
  offset,
  transactionCount,
  total,
  onPageChange,
  onLimitChange,
}: Props) {
  return (
    <div className="rounded-2xl bg-white p-4 shadow-[0_1px_3px_rgba(15,23,42,0.04)]">
      <div className="flex items-center justify-between gap-3">
        <span className="text-xs text-[#737686]">
          Showing{" "}
          <strong className="text-[#0b1c30]">
            {offset + 1}-{Math.min(offset + transactionCount, total)}
          </strong>{" "}
          of{" "}
          <strong className="text-[#0b1c30]">{total}</strong>
        </span>

        <select
          value={limit}
          onChange={(event) =>
            onLimitChange(Number(event.target.value))
          }
          className="rounded-lg bg-[#eff4ff] px-2 py-1 text-xs font-semibold outline-none"
        >
          {PAGE_SIZE_OPTIONS.map((size) => (
            <option key={size} value={size}>
              {size}
            </option>
          ))}
        </select>
      </div>

      <div className="mt-4 flex items-center justify-between">
        <button
          type="button"
          disabled={page === 1}
          onClick={() => onPageChange(Math.max(1, page - 1))}
          className="rounded-lg bg-[#eff4ff] px-3 py-2 text-xs font-semibold disabled:cursor-not-allowed disabled:opacity-40"
        >
          Previous
        </button>

        <span className="text-xs font-semibold text-[#737686]">
          Page {page} of {totalPages}
        </span>

        <button
          type="button"
          disabled={page >= totalPages}
          onClick={() =>
            onPageChange(Math.min(totalPages, page + 1))
          }
          className="rounded-lg bg-[#eff4ff] px-3 py-2 text-xs font-semibold disabled:cursor-not-allowed disabled:opacity-40"
        >
          Next
        </button>
      </div>
    </div>
  );
}
