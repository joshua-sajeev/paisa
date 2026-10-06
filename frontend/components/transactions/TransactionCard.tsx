"use client";

import {
  AccountIcon,
  formatMoney,
  formatTime,
  getTypeClasses,
  TransactionIcon,
} from "./transaction-utils";

import type { Transaction } from "./transaction-utils";

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

  return (
    <article className="rounded-2xl border border-black/[0.04] bg-white p-4 shadow-[0_1px_3px_rgba(15,23,42,0.04)] transition hover:shadow-sm">
      <div className="flex items-start justify-between gap-3">
        <div className="flex min-w-0 items-start gap-3">
          <div
            className={`flex h-11 w-11 shrink-0 items-center justify-center rounded-full ${styles.icon}`}
          >
            <TransactionIcon
              type={transaction.type}
              category={transaction.category}
            />
          </div>

          <div className="min-w-0">
            <h3 className="truncate text-sm font-semibold text-[#0b1c30]">
              {transaction.name}
            </h3>

            <div className="mt-1.5 flex flex-wrap items-center gap-2">
              <span className="inline-flex items-center gap-1 rounded-lg bg-[#e5eeff] px-2 py-1 text-[11px] font-semibold text-[#434655]">
                <AccountIcon account={transaction.account} />
                {transaction.account}
              </span>

              <span className="text-xs text-[#737686]">
                {formatTime(transaction.occurred_at)}
              </span>

              {transaction.jar_name && (
                <span className="rounded-lg bg-[#eff4ff] px-2 py-1 text-[11px] font-medium text-[#434655]">
                  {transaction.jar_name}
                </span>
              )}

              <span className="rounded-lg bg-[#f5f5f7] px-2 py-1 text-[11px] font-medium capitalize text-[#737686]">
                {transaction.category}
              </span>
            </div>
          </div>
        </div>

        <div className="shrink-0 text-right">
          <p className={`text-sm font-bold tracking-tight ${styles.amount}`}>
            {amountPrefix}
            {formatMoney(transaction.amount)}
          </p>
        </div>
      </div>
    </article>
  );
}
