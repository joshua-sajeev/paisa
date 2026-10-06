import TransactionCard from "./TransactionCard";
import type { Transaction } from "./transaction-utils";

type Props = {
  group: string;
  transactions: Transaction[];
};

export default function TransactionGroup({
  group,
  transactions,
}: Props) {
  return (
    <section className="flex flex-col gap-2.5">
      <div className="flex items-center justify-between px-1">
        <div className="flex items-center gap-2">
          <h2 className="text-base font-semibold text-[#0b1c30]">
            {group}
          </h2>

          {group !== "Today" && group !== "Yesterday" && (
            <span className="rounded-full bg-[#e5eeff] px-2 py-0.5 text-[10px] font-bold text-[#737686]">
              {transactions.length}
            </span>
          )}
        </div>

        <span className="text-xs text-[#737686]">
          {transactions.length}{" "}
          {transactions.length === 1 ? "entry" : "entries"}
        </span>
      </div>

      <div className="flex flex-col gap-2.5">
        {transactions.map((transaction) => (
          <TransactionCard
            key={transaction.id}
            transaction={transaction}
          />
        ))}
      </div>
    </section>
  );
}
