"use client";

import {
  formatMoney,
  formatTime,
  getTypeClasses,
  TransactionIcon,
} from "./transaction-utils";

import type { Transaction } from "./transaction-utils";

const labelColors = ["#6366F1", "#07B682", "#F59E0B", "#EC4899", "#8B5CF6"];

function getLabelColor(label: string): string {
  let hash = 0;
  for (let index = 0; index < label.length; index++) {
    hash = (hash * 31 + label.charCodeAt(index)) >>> 0;
  }
  return labelColors[hash % labelColors.length];
}

type Props = {
  transaction: Transaction;
};

export default function TransactionCard({ transaction }: Props) {
  const styles = getTypeClasses(transaction.type);

  const amountPrefix =
    transaction.type === "income"
      ? "+"
      : transaction.type === "expense"
        ? "-"
        : "";

  const category = transaction.category
    ? transaction.category.charAt(0).toUpperCase() + transaction.category.slice(1)
    : "Uncategorized";
  const accountColor = getLabelColor(transaction.account);

  return (
    <div className="flex items-center gap-3 px-3.5 py-3 transition-colors hover:bg-slate-50/70">
      <div
        className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-xl ${styles.icon}`}
      >
        <TransactionIcon
          type={transaction.type}
          category={transaction.category}
        />
      </div>

      <div className="min-w-0 flex-1">
        <div className="flex items-center justify-between gap-2">
          <div className="flex min-w-0 items-center gap-1.5">
            <h3 className="truncate text-[14px] font-bold leading-tight text-[#0b1c30]">
              {transaction.name}
            </h3>
            <span
              className="shrink-0 rounded-md border px-1.5 py-0.5 text-[11px] font-semibold"
              style={{
                color: accountColor,
                backgroundColor: `${accountColor}1A`,
                borderColor: `${accountColor}33`,
              }}
            >
              {transaction.account}
            </span>
          </div>

          <p className={`shrink-0 text-sm font-bold tracking-tight ${styles.amount}`}>
            {amountPrefix}
            {formatMoney(transaction.amount)}
          </p>
        </div>

        <div className="mt-0.5 flex min-w-0 items-center justify-between gap-2 text-[11px] text-[#737686]">
          <p className="min-w-0 truncate">
            {category}
            {transaction.jar_name && <span> • {transaction.jar_name}</span>}
          </p>
          <span className="ml-auto shrink-0 text-right text-[#737686]/70">
            {formatTime(transaction.occurred_at)}
          </span>
        </div>
      </div>
    </div>
  );
}
