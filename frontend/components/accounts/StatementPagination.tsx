"use client";

type Props = {
  page: number;
  totalPages: number;
  total: number;
  pageSize: number;
  onPageChange: (page: number) => void;
};

export default function StatementPagination({
  page,
  totalPages,
  total,
  pageSize,
  onPageChange,
}: Props) {
  if (total === 0) {
    return null;
  }

  const start = (page - 1) * pageSize + 1;
  const end = Math.min(page * pageSize, total);

  return (
    <section className="mt-2 flex items-center justify-between rounded-2xl border border-amber-900/5 bg-white p-3 shadow-sm">
      <span className="text-xs font-semibold text-slate-500">
        Showing{" "}
        <span className="font-bold text-slate-800">
          {start}–{end}
        </span>{" "}
        of{" "}
        <span className="font-bold text-slate-800">
          {total}
        </span>
      </span>

      <div className="flex items-center gap-1.5">
        <button
          type="button"
          disabled={page <= 1}
          onClick={() => onPageChange(page - 1)}
          className="rounded-lg bg-slate-100 px-2.5 py-1 text-xs font-semibold text-slate-400 transition-colors disabled:cursor-not-allowed enabled:text-slate-700 enabled:hover:bg-slate-200"
        >
          Previous
        </button>

        <span className="rounded-md bg-blue-50 px-2 py-0.5 text-xs font-bold text-blue-600">
          {page}
        </span>

        <button
          type="button"
          disabled={page >= totalPages}
          onClick={() => onPageChange(page + 1)}
          className="rounded-lg border border-amber-900/10 bg-[#fffaf0] px-2.5 py-1 text-xs font-semibold text-slate-700 transition-colors hover:bg-amber-100/60 disabled:cursor-not-allowed disabled:opacity-40"
        >
          Next
        </button>
      </div>
    </section>
  );
}
