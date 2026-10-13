import TransactionCard from "./TransactionCard";
import { formatMoney, type Transaction } from "./transaction-utils";

type Props = {
  group: string;
  transactions: Transaction[];
  onDelete: (transaction: Transaction) => void;
};

export default function TransactionGroup({
  group,
  transactions,
  onDelete,
}: Props) {
  const net = transactions.reduce((total, transaction) => {
    if (transaction.type === "income") return total + transaction.amount;
    if (transaction.type === "expense") return total - transaction.amount;
    return total;
  }, 0);

  return (
    <section className="space-y-1.5">
      <div className="flex items-center justify-between px-1">
        <div className="flex items-center gap-2">
          <h2 className="text-sm font-semibold text-[#0b1c30]">
            {group}
          </h2>

          {group !== "Today" && group !== "Yesterday" && (
            <span className="rounded-full bg-[#e5eeff] px-2 py-0.5 text-[10px] font-bold text-[#737686]">
              {transactions.length}
            </span>
          )}
        </div>

        <span
          className={`text-xs font-semibold ${net >= 0 ? "text-emerald-600" : "text-rose-500"}`}
        >
          Net: {net >= 0 ? "+" : "-"}
          {formatMoney(Math.abs(net))}
        </span>
      </div>

      <div className="divide-y divide-slate-100 overflow-hidden rounded-2xl border border-amber-900/5 bg-white shadow-sm">
        {transactions.map((transaction) => (
          <TransactionCard
            key={transaction.id}
            transaction={transaction}
            onDelete={onDelete}
          />
        ))}
      </div>
    </section>
  );
}
