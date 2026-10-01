"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";

import {
  ArrowBack,
  ArrowDownward,
  ArrowUpward,
  CalendarToday,
  Search,
  AccountBalance,
  Payments,
  ShoppingCart,
  Restaurant,
  Savings,
  VolunteerActivism,
  LocalPharmacy,
  Bolt,
} from "@material-symbols-svg/react/w400";

type Transaction = {
  id: string;
  name: string;
  date: string;
  type: "income" | "expense" | "transfer";
  amount: number;
  jar_name: string | null;
  account: string;
  account_balance: number;
  category: string;
};

type TransactionsResponse = {
  transactions: Transaction[];
  total: number;
};

type Props = {
  accountId: string;
};

const API_URL =
  process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

function formatMoney(amount: number) {
  return new Intl.NumberFormat("en-IN", {
    style: "currency",
    currency: "INR",
    minimumFractionDigits: 2,
  }).format(amount / 100);
}

function formatDate(date: string) {
  return new Intl.DateTimeFormat("en-IN", {
    day: "numeric",
    month: "short",
    year: "numeric",
  }).format(new Date(`${date}T00:00:00`));
}

function getIcon(
  category: string,
  type: Transaction["type"],
) {
  if (type === "income") {
    return Payments;
  }

  switch (category) {
    case "food":
      return Restaurant;

    case "groceries":
      return ShoppingCart;

    case "health":
      return LocalPharmacy;

    case "investment":
      return Savings;

    case "donation":
      return VolunteerActivism;

    case "housing":
      return Bolt;

    case "transfer":
      return AccountBalance;

    default:
      return ShoppingCart;
  }
}

export default function AccountStatement({
  accountId,
}: Props) {
  const router = useRouter();

  const [transactions, setTransactions] = useState<
    Transaction[]
  >([]);

  const [total, setTotal] = useState(0);

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const [search, setSearch] = useState("");
  const [filter, setFilter] = useState<
    "all" | "income" | "expense" | "transfer"
  >("all");

  useEffect(() => {
    async function fetchTransactions() {
      try {
        setLoading(true);
        setError("");

        const response = await fetch(
          `${API_URL}/api/v1/accounts/${accountId}/transactions/`,
          {
            headers: {
              "Content-Type": "application/json",
            },
            credentials: "include",
          },
        );

        if (!response.ok) {
          throw new Error(
            "Failed to fetch transactions",
          );
        }

        const data: TransactionsResponse =
          await response.json();

        setTransactions(data.transactions ?? []);
        setTotal(data.total ?? 0);
      } catch {
        setError(
          "Failed to load account statement.",
        );
      } finally {
        setLoading(false);
      }
    }

    fetchTransactions();
  }, [accountId]);

  const filteredTransactions = useMemo(() => {
    const query = search.trim().toLowerCase();

    return transactions.filter((transaction) => {
      const matchesSearch =
        !query ||
        transaction.name
          .toLowerCase()
          .includes(query) ||
        transaction.category
          .toLowerCase()
          .includes(query);

      const matchesFilter =
        filter === "all" ||
        transaction.type === filter;

      return matchesSearch && matchesFilter;
    });
  }, [transactions, search, filter]);

  const groups = useMemo(() => {
    const grouped = new Map<
      string,
      Transaction[]
    >();

    for (const transaction of filteredTransactions) {
      const current =
        grouped.get(transaction.date) ?? [];

      current.push(transaction);

      grouped.set(transaction.date, current);
    }

    return Array.from(grouped.entries());
  }, [filteredTransactions]);

  const income = transactions
    .filter((transaction) => transaction.type === "income")
    .reduce(
      (sum, transaction) =>
        sum + transaction.amount,
      0,
    );

  const expenses = transactions
    .filter(
      (transaction) => transaction.type === "expense",
    )
    .reduce(
      (sum, transaction) =>
        sum + Math.abs(transaction.amount),
      0,
    );

  const net = income - expenses;

  return (
    <main className="min-h-screen bg-[#fffaf0] pb-8 text-slate-800">
      {/* Header */}
      <header className="sticky top-0 z-40 border-b border-amber-900/5 bg-[#fffaf0]/90 backdrop-blur-md">
        <div className="mx-auto flex h-14 max-w-md items-center justify-between px-4">
          <button
            type="button"
            onClick={() => router.back()}
            className="inline-flex items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-xs font-semibold text-slate-600 transition hover:bg-black/5 hover:text-slate-900"
          >
            <ArrowBack
              width={18}
              height={18}
            />

            <span>Back to Accounts</span>
          </button>

          <h1 className="text-sm font-bold tracking-tight text-slate-900">
            Statement
          </h1>

          <div className="w-20" />
        </div>
      </header>

      <div className="mx-auto w-full max-w-md space-y-4 px-4 pt-3">
        {/* Summary */}
        <section className="rounded-2xl border border-amber-900/5 bg-white p-4 shadow-sm">
          <div className="flex items-center gap-2.5">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-blue-600 text-white">
              <AccountBalance
                width={22}
                height={22}
              />
            </div>

            <div>
              <p className="text-sm font-bold text-slate-900">
                Account Statement
              </p>

              <p className="text-[11px] text-slate-400">
                {total} transactions
              </p>
            </div>
          </div>

          <div className="mt-4 grid grid-cols-3 gap-2 border-t border-slate-100 pt-3">
            <div className="rounded-xl bg-[#fffaf0] p-2.5">
              <div className="flex items-center gap-1 text-[10px] font-bold uppercase text-slate-400">
                <ArrowDownward
                  width={13}
                  height={13}
                />
                Inflow
              </div>

              <p className="mt-1 text-[13px] font-bold text-[#07B682]">
                +{formatMoney(income)}
              </p>
            </div>

            <div className="rounded-xl bg-[#fffaf0] p-2.5">
              <div className="flex items-center gap-1 text-[10px] font-bold uppercase text-slate-400">
                <ArrowUpward
                  width={13}
                  height={13}
                />
                Outflow
              </div>

              <p className="mt-1 text-[13px] font-bold text-[#F43F5E]">
                -{formatMoney(expenses)}
              </p>
            </div>

            <div className="rounded-xl bg-[#fffaf0] p-2.5">
              <p className="text-[10px] font-bold uppercase text-slate-400">
                Net
              </p>

              <p
                className={`mt-1 text-[13px] font-bold ${
                  net >= 0
                    ? "text-blue-600"
                    : "text-[#F43F5E]"
                }`}
              >
                {net >= 0 ? "+" : ""}
                {formatMoney(net)}
              </p>
            </div>
          </div>
        </section>

        {/* Filters */}
        <section className="space-y-2.5 rounded-2xl border border-amber-900/5 bg-white p-3 shadow-sm">
          <div className="relative">
            <Search
              width={18}
              height={18}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"
            />

            <input
              value={search}
              onChange={(event) =>
                setSearch(event.target.value)
              }
              placeholder="Search transactions..."
              className="w-full rounded-xl border border-amber-900/10 bg-[#fffaf0] py-2 pl-9 pr-4 text-xs font-medium outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20"
            />
          </div>

          <div className="flex gap-1.5 overflow-x-auto">
            {[
              ["all", `All (${total})`],
              ["income", "Income"],
              ["expense", "Expense"],
              ["transfer", "Transfers"],
            ].map(([value, label]) => (
              <button
                key={value}
                type="button"
                onClick={() =>
                  setFilter(
                    value as
                      | "all"
                      | "income"
                      | "expense"
                      | "transfer",
                  )
                }
                className={`shrink-0 rounded-full px-3 py-1 text-xs font-bold ${
                  filter === value
                    ? "bg-slate-900 text-white"
                    : "border border-amber-900/10 bg-[#fffaf0] text-slate-600"
                }`}
              >
                {label}
              </button>
            ))}
          </div>

          <button
            type="button"
            className="inline-flex items-center gap-1.5 rounded-lg border border-amber-900/10 bg-[#fffaf0] px-2.5 py-1.5 text-xs font-semibold text-slate-700"
          >
            <CalendarToday
              width={16}
              height={16}
              className="text-blue-600"
            />
            All dates
          </button>
        </section>

        {/* Loading */}
        {loading && (
          <div className="rounded-2xl bg-white p-8 text-center text-xs font-semibold text-slate-400">
            Loading statement...
          </div>
        )}

        {/* Error */}
        {!loading && error && (
          <div className="rounded-2xl bg-red-50 p-4 text-center text-xs font-semibold text-red-600">
            {error}
          </div>
        )}

        {/* Transactions */}
        {!loading &&
          !error &&
          groups.map(([date, items]) => {
            const groupNet = items.reduce(
              (sum, transaction) =>
                sum + transaction.amount,
              0,
            );

            return (
              <section
                key={date}
                className="space-y-1.5"
              >
                <div className="flex items-center justify-between px-1">
                  <h2 className="text-xs font-bold uppercase tracking-wider text-slate-500">
                    {formatDate(date)}
                  </h2>

                  <span
                    className={`text-[11px] font-semibold ${
                      groupNet >= 0
                        ? "text-emerald-600"
                        : "text-rose-500"
                    }`}
                  >
                    Net:{" "}
                    {groupNet >= 0 ? "+" : ""}
                    {formatMoney(groupNet)}
                  </span>
                </div>

                <div className="divide-y divide-slate-100 overflow-hidden rounded-2xl border border-amber-900/5 bg-white shadow-sm">
                  {items.map((transaction) => {
                    const Icon = getIcon(
                      transaction.category,
                      transaction.type,
                    );

                    const positive =
                      transaction.amount > 0;

                    return (
                      <div
                        key={transaction.id}
                        className="flex items-center justify-between gap-3 p-3.5"
                      >
                        <div className="flex min-w-0 items-center gap-3">
                          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border border-slate-100 bg-slate-50 text-slate-600">
                            <Icon
                              width={20}
                              height={20}
                            />
                          </div>

                          <div className="min-w-0">
                            <p className="truncate text-xs font-bold text-slate-900">
                              {transaction.name}
                            </p>

                            <div className="mt-0.5 flex items-center gap-1.5 text-[11px] text-slate-500">
                              <span>
                                {transaction.category}
                              </span>

                              {transaction.jar_name && (
                                <>
                                  <span>•</span>
                                  <span>
                                    {transaction.jar_name}
                                  </span>
                                </>
                              )}
                            </div>
                          </div>
                        </div>

                        <div className="shrink-0 text-right">
                          <p
                            className={`text-xs font-bold ${
                              positive
                                ? "text-[#07B682]"
                                : "text-[#F43F5E]"
                            }`}
                          >
                            {positive ? "+" : ""}
                            {formatMoney(
                              transaction.amount,
                            )}
                          </p>

                          <p className="mt-0.5 text-[10px] text-slate-400">
                            Bal:{" "}
                            {formatMoney(
                              transaction.account_balance,
                            )}
                          </p>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </section>
            );
          })}
      </div>
    </main>
  );
}
