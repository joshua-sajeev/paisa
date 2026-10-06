"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";

import StatementHeader from "./StatementHeader";
import StatementFilters, {
  type DateRange,
  isInDateRange,
} from "./StatementFilters";
import StatementTransactionList from "./StatementTransactionList";
import StatementPagination from "./StatementPagination";

export type StatementTransaction = {
  id: string;
  name: string;
  occurred_at: string;
  type: "income" | "expense" | "transfer";
  category: string;
  jar_name: string | null;
  amount: number;
  balance_after: number;
};

export type StatementAccount = {
  id: string;
  name: string;
  icon_key: string | null;
  balance: number;
  is_primary: boolean;
  is_archived: boolean;
};

type StatementResponse = {
  account: StatementAccount;
  transactions: StatementTransaction[];
  total: number;
};

const PAGE_SIZE = 8;

export default function AccountStatement({
  accountId,
}: {
  accountId: string;
}) {
  const [account, setAccount] = useState<StatementAccount | null>(null);
  const [transactions, setTransactions] = useState<StatementTransaction[]>(
    [],
  );
  const [search, setSearch] = useState("");
  const [filter, setFilter] = useState<
    "all" | "income" | "expense" | "transfer"
  >("all");
  const [dateRange, setDateRange] = useState<DateRange>(null);

  const [jarFilter, setJarFilter] = useState("all");
  const [categoryFilter, setCategoryFilter] = useState("all");
  const [minAmount, setMinAmount] = useState("");
  const [maxAmount, setMaxAmount] = useState("");
  const [page, setPage] = useState(1);

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let cancelled = false;

    async function loadStatement() {
      setLoading(true);
      setError("");

      try {
        const response = await fetch(
          `/api/v1/accounts/${accountId}/statement`,
          {
            method: "GET",
            headers: {
              Accept: "application/json",
            },
            cache: "no-store",
          },
        );

        if (!response.ok) {
          throw new Error(
            `Failed to load statement (${response.status})`,
          );
        }

        const data: StatementResponse = await response.json();

        if (cancelled) {
          return;
        }

        setAccount(data.account);
        setTransactions(data.transactions ?? []);
        setPage(1);
      } catch (err) {
        if (cancelled) {
          return;
        }

        console.error("Failed to load account statement:", err);
        setError("Failed to load account statement.");
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    loadStatement();

    return () => {
      cancelled = true;
    };
  }, [accountId]);

  const filteredTransactions = useMemo(() => {
    const query = search.trim().toLowerCase();
    const min = minAmount ? parseFloat(minAmount) : null;
    const max = maxAmount ? parseFloat(maxAmount) : null;

    return transactions
      .filter((transaction) => {
        if (filter !== "all" && transaction.type !== filter) {
          return false;
        }

        if (jarFilter !== "all" && transaction.jar_name !== jarFilter) {
          return false;
        }

        if (
          categoryFilter !== "all" &&
          transaction.category !== categoryFilter
        ) {
          return false;
        }

        const absAmount = Math.abs(transaction.amount);
        if (min !== null && !isNaN(min) && absAmount < min) {
          return false;
        }
        if (max !== null && !isNaN(max) && absAmount > max) {
          return false;
        }

        if (!isInDateRange(transaction.occurred_at, dateRange)) {
          return false;
        }

        if (!query) {
          return true;
        }

        return (
          transaction.name.toLowerCase().includes(query) ||
          transaction.category.toLowerCase().includes(query) ||
          transaction.jar_name?.toLowerCase().includes(query)
        );
      })
      .sort(
        (a, b) =>
          new Date(b.occurred_at).getTime() -
          new Date(a.occurred_at).getTime(),
      );
  }, [
    transactions,
    search,
    filter,
    jarFilter,
    categoryFilter,
    minAmount,
    maxAmount,
    dateRange,
  ]);

  function handleJarFilter(value: string) {
    setJarFilter(value);
    setPage(1);
  }

  function handleCategoryFilter(value: string) {
    setCategoryFilter(value);
    setPage(1);
  }

  function handleMinAmount(value: string) {
    setMinAmount(value);
    setPage(1);
  }

  function handleMaxAmount(value: string) {
    setMaxAmount(value);
    setPage(1);
  }

  const totalPages = Math.max(
    1,
    Math.ceil(filteredTransactions.length / PAGE_SIZE),
  );

  const currentPage = Math.min(page, totalPages);
  const paginatedTransactions = useMemo(() => {
    const start = (currentPage - 1) * PAGE_SIZE;
    return filteredTransactions.slice(start, start + PAGE_SIZE);
  }, [filteredTransactions, currentPage]);

  const summary = useMemo(() => {
    let inflow = 0;
    let outflow = 0;

    for (const transaction of transactions) {
      if (!isInDateRange(transaction.occurred_at, dateRange)) {
        continue;
      }

      if (
        transaction.type === "income" ||
        (transaction.type === "transfer" && transaction.amount > 0)
      ) {
        inflow += Math.abs(transaction.amount);
      } else if (
        transaction.type === "expense" ||
        (transaction.type === "transfer" && transaction.amount < 0)
      ) {
        outflow += Math.abs(transaction.amount);
      }
    }

    return {
      inflow,
      outflow,
      net: inflow - outflow,
    };
  }, [transactions, dateRange]);

  function handleSearch(value: string) {
    setSearch(value);
    setPage(1);
  }

  function handleFilter(
    value: "all" | "income" | "expense" | "transfer",
  ) {
    setFilter(value);
    setPage(1);
  }

  function handleDateRange(value: DateRange) {
    setDateRange(value);
    setPage(1);
  }

  if (loading) {
    return (
      <main className="min-h-screen bg-[#fffaf0]">
        <div className="mx-auto min-h-screen w-full max-w-md px-4 py-5">
          <div className="animate-pulse space-y-4">
            <div className="h-8 w-40 rounded-xl bg-black/5" />
            <div className="h-56 rounded-3xl bg-black/5" />
            <div className="h-12 rounded-2xl bg-black/5" />
            <div className="h-12 rounded-2xl bg-black/5" />
            <div className="h-24 rounded-2xl bg-black/5" />
            <div className="h-24 rounded-2xl bg-black/5" />
            <div className="h-24 rounded-2xl bg-black/5" />
          </div>
        </div>
      </main>
    );
  }

  if (error || !account) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-[#fffaf0] px-4">
        <div className="text-center">
          <p className="text-sm font-semibold text-error">
            {error || "Account not found."}
          </p>

          <Link
            href="/accounts"
            className="mt-4 inline-flex rounded-xl bg-black px-4 py-2 text-sm font-semibold text-white"
          >
            Back to Accounts
          </Link>
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-[#fffaf0] ">
      <div className="mx-auto w-full max-w-md px-4">
        <StatementHeader
          account={account}
          inflow={summary.inflow}
          outflow={summary.outflow}
          net={summary.net}
        />

        <StatementFilters
          search={search}
          filter={filter}
          jarFilter={jarFilter}
          categoryFilter={categoryFilter}
          minAmount={minAmount}
          maxAmount={maxAmount}
          dateRange={dateRange}
          transactions={transactions}
          onSearchChange={handleSearch}
          onFilterChange={handleFilter}
          onJarFilterChange={handleJarFilter}
          onCategoryFilterChange={handleCategoryFilter}
          onMinAmountChange={handleMinAmount}
          onMaxAmountChange={handleMaxAmount}
          onDateRangeChange={handleDateRange}
        />

        <StatementTransactionList transactions={paginatedTransactions} />

        <StatementPagination
          page={currentPage}
          totalPages={totalPages}
          total={filteredTransactions.length}
          pageSize={PAGE_SIZE}
          onPageChange={setPage}
        />
      </div>
    </main>
  );
}
