
"use client";

import type { ComponentType, SVGProps } from "react";

import { DirectionsCarFillW700 } from "@material-symbols-svg/react/icons/directions-car";
import { HomeFillW700 } from "@material-symbols-svg/react/icons/home";
import { LocalPharmacyFillW700 } from "@material-symbols-svg/react/icons/local-pharmacy";
import { MovieFillW700 } from "@material-symbols-svg/react/icons/movie";
import { PaymentsFillW700 } from "@material-symbols-svg/react/icons/payments";
import { ReceiptLongFillW700 } from "@material-symbols-svg/react/icons/receipt-long";
import { RestaurantFillW700 } from "@material-symbols-svg/react/icons/restaurant";
import { ShoppingCartFillW700 } from "@material-symbols-svg/react/icons/shopping-cart";
import { SwapHorizFillW700 } from "@material-symbols-svg/react/icons/swap-horiz";
import { TrendingUpFillW700 } from "@material-symbols-svg/react/icons/trending-up";
import { VolunteerActivismFillW700 } from "@material-symbols-svg/react/icons/volunteer-activism";

import type { StatementTransaction } from "./AccountStatement";

type Props = {
  transactions: StatementTransaction[];
};

type IconComponent = ComponentType<SVGProps<SVGSVGElement>>;

const CATEGORY_ICONS: Record<string, IconComponent> = {
  food: RestaurantFillW700,
  groceries: ShoppingCartFillW700,
  entertainment: MovieFillW700,
  transport: DirectionsCarFillW700,
  health: LocalPharmacyFillW700,
  investment: TrendingUpFillW700,
  housing: HomeFillW700,
  donation: VolunteerActivismFillW700,
  other: ReceiptLongFillW700,
};

function formatMoney(paise: number) {
  return new Intl.NumberFormat("en-IN", {
    style: "currency",
    currency: "INR",
    maximumFractionDigits: 2,
  }).format(Math.abs(paise) / 100);
}

function formatDate(dateString: string) {
  const date = new Date(dateString);

  return new Intl.DateTimeFormat("en-IN", {
    day: "numeric",
    month: "short",
    year: "numeric",
  }).format(date);
}

function formatTime(dateString: string) {
  const date = new Date(dateString);

  return new Intl.DateTimeFormat("en-IN", {
    hour: "numeric",
    minute: "2-digit",
  }).format(date);
}

function dateKey(dateString: string) {
  const date = new Date(dateString);

  return new Intl.DateTimeFormat("en-CA", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(date);
}

function groupLabel(dateString: string) {
  const today = new Date();

  if (dateKey(dateString) === dateKey(today.toISOString())) {
    return "Today";
  }

  return formatDate(dateString);
}

function formatCategory(category: string) {
  return category.charAt(0).toUpperCase() + category.slice(1);
}

function TransactionIcon({
  category,
  type,
  className,
}: {
  category: string;
  type: string;
  className?: string;
}) {
  if (type === "income") {
    return <PaymentsFillW700 className={className} />;
  }

  if (type === "transfer") {
    return <SwapHorizFillW700 className={className} />;
  }

  const Icon = CATEGORY_ICONS[category] ?? ReceiptLongFillW700;

  return <Icon className={className} />;
}

export default function StatementTransactionList({
  transactions,
}: Props) {
  if (transactions.length === 0) {
    return (
      <section className="rounded-2xl border border-amber-900/5 bg-white px-4 py-12 text-center shadow-sm">
        <ReceiptLongFillW700 className="mx-auto h-8 w-8 text-slate-300" />

        <p className="mt-2 text-sm font-bold text-slate-700">
          No transactions found
        </p>

        <p className="mt-1 text-xs text-slate-400">
          Try changing your search or filter.
        </p>
      </section>
    );
  }

  const groups = groupTransactions(transactions);

  return (
    <div className="space-y-4">
      {groups.map((group) => (
        <section key={group.key} className="space-y-1.5">
          <div className="flex items-center justify-between px-1">
            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-500">
              {groupLabel(group.transactions[0].occurred_at)}
            </h3>

            <span
              className={[
                "text-[11px] font-semibold",
                group.net >= 0
                  ? "text-emerald-600"
                  : "text-rose-500",
              ].join(" ")}
            >
              Net: {group.net >= 0 ? "+" : "-"}
              {formatMoney(group.net)}
            </span>
          </div>

          <div className="divide-y divide-slate-100 overflow-hidden rounded-2xl border border-amber-900/5 bg-white shadow-sm">
            {group.transactions.map((transaction) => (
              <StatementTransactionRow
                key={transaction.id}
                transaction={transaction}
              />
            ))}
          </div>
        </section>
      ))}
    </div>
  );
}
// NOTE: keep your existing imports here (StatementTransaction, TransactionIcon,
// formatTime, formatCategory, formatMoney, dateKey, etc.)

const JAR_COLORS = [
  "#6366F1",
  "#07B682",
  "#F59E0B",
  "#EC4899",
  "#8B5CF6",
];

// Hash the jar name into a stable palette index so the same jar
// always gets the same color across rows and renders.
function getJarColor(jarName: string): string {
  let hash = 0;
  for (let i = 0; i < jarName.length; i++) {
    hash = (hash * 31 + jarName.charCodeAt(i)) >>> 0;
  }
  return JAR_COLORS[hash % JAR_COLORS.length];
}

function StatementTransactionRow({
  transaction,
}: {
  transaction: StatementTransaction;
}) {
  const isIncome = transaction.type === "income";
  const isTransfer = transaction.type === "transfer";

  const amountClass = isIncome
    ? "text-[#07B682]"
    : isTransfer
      ? "text-blue-700"
      : "text-[#F43F5E]";

  const iconClass = isIncome
    ? "bg-emerald-50 text-emerald-600 border-emerald-100"
    : isTransfer
      ? "bg-blue-50 text-blue-600 border-blue-100"
      : "bg-rose-50 text-rose-600 border-rose-100";

  const amountPrefix = isIncome ? "+" : isTransfer ? "" : "-";

  const jarColor = transaction.jar_name
    ? getJarColor(transaction.jar_name)
    : null;

  return (
    <div className="flex items-center justify-between gap-3 p-3.5 transition-colors hover:bg-slate-50/70">
      <div className="flex min-w-0 items-center gap-3">
        <div
          className={`flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl border ${iconClass}`}
        >
          <TransactionIcon
            category={transaction.category}
            type={transaction.type}
            className="h-5 w-5"
          />
        </div>

        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <p className="truncate text-xs font-bold text-slate-900">
              {transaction.name}
            </p>

            {transaction.jar_name && jarColor && (
              <span
                className="inline-flex flex-shrink-0 items-center rounded border px-1.5 py-0.5 text-[10px] font-semibold"
                style={{
                  color: jarColor,
                  backgroundColor: `${jarColor}1A`, // ~10% opacity
                  borderColor: `${jarColor}33`, // ~20% opacity
                }}
              >
                {transaction.jar_name}
              </span>
            )}
          </div>

          <div className="mt-0.5 flex items-center gap-1.5 text-[11px] text-slate-500">
            <span>{formatTime(transaction.occurred_at)}</span>

            <span>•</span>

            <span className="font-medium text-slate-600">
              {formatCategory(transaction.category)}
            </span>
          </div>
        </div>
      </div>

      <div className="flex-shrink-0 text-right">
        <div className={`text-xs font-bold ${amountClass}`}>
          {amountPrefix}
          {formatMoney(transaction.amount)}
        </div>

        <div className="mt-0.5 text-[10px] text-slate-400">
          Bal: {formatMoney(transaction.balance_after)}
        </div>
      </div>
    </div>
  );
}

type TransactionGroup = {
  key: string;
  transactions: StatementTransaction[];
  net: number;
};

function groupTransactions(
  transactions: StatementTransaction[],
): TransactionGroup[] {
  const groups = new Map<string, StatementTransaction[]>();

  const sorted = [...transactions].sort(
    (a, b) =>
      new Date(b.occurred_at).getTime() -
      new Date(a.occurred_at).getTime(),
  );

  for (const transaction of sorted) {
    const key = dateKey(transaction.occurred_at);

    const existing = groups.get(key);

    if (existing) {
      existing.push(transaction);
    } else {
      groups.set(key, [transaction]);
    }
  }

  return Array.from(groups.entries()).map(([key, group]) => ({
    key,
    transactions: group,
    net: group.reduce(
      (total, transaction) => total + transaction.amount,
      0,
    ),
  }));
}
