"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "next/navigation";

import { apiFetch } from "@/lib/api";
import type { DateRange } from "@/components/ui/DateRangePicker";

import TransactionsFilters from "./TransactionsFilters";
import TransactionGroup from "./TransactionGroup";
import TransactionsPagination from "./TransactionsPagination";
import TransactionsSkeleton from "./TransactionsSkeleton";
import TransactionsEmpty from "./TransactionsEmpty";

import {
  groupTransactions,
  type Account,
  type Jar,
  type Transaction,
  type TransactionCategory,
  type TransactionType,
  type TransactionsResponse,
} from "./transaction-utils";

export default function TransactionsList() {
  const searchParams = useSearchParams();
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [total, setTotal] = useState(0);

  const [accounts, setAccounts] = useState<Account[]>([]);
  const [jars, setJars] = useState<Jar[]>([]);

  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(5);

  const [loading, setLoading] = useState(true);
  const [loadingFilters, setLoadingFilters] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [search, setSearch] = useState(() => searchParams.get("search") ?? "");
  const [type, setType] =
    useState<TransactionType | "all">(() => (searchParams.get("type") as TransactionType | "all") ?? "all");
  const [account, setAccount] = useState(() => searchParams.get("account") ?? searchParams.get("account_id") ?? "all");
  const [jar, setJar] = useState(() => searchParams.get("jar") ?? searchParams.get("jar_id") ?? "all");
  const [category, setCategory] =
    useState<TransactionCategory | "all">(() => (searchParams.get("category") as TransactionCategory | "all") ?? "all");
  const [dateRange, setDateRange] =
    useState<DateRange>(() => {
      const start = searchParams.get("from_date") ?? searchParams.get("from");
      const end = searchParams.get("to_date") ?? searchParams.get("to");
      return start && end ? { start, end } : null;
    });

  const [filterOpen, setFilterOpen] = useState(false);

  const offset = (page - 1) * limit;

  const totalPages = Math.max(1, Math.ceil(total / limit));

  // load accounts/jars
  useEffect(() => {
    async function loadFilters() {
      setLoadingFilters(true);

      try {
        const [accountsData, jarsData] = await Promise.all([
          apiFetch<Account[]>("/api/v1/accounts"),
          apiFetch<Jar[]>("/api/v1/jars"),
        ]);

        setAccounts(
          accountsData.filter(
            (item) => !item.is_archived,
          ),
        );

        setJars(
          jarsData.filter(
            (item) => !item.is_archived,
          ),
        );
      } catch (err) {
        console.error("Failed to load transaction filters", err);
      } finally {
        setLoadingFilters(false);
      }
    }

    loadFilters();
  }, []);

  // fetch transactions
  const fetchTransactions = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const params = new URLSearchParams();

      params.set("limit", String(limit));
      params.set("offset", String(offset));

      if (type !== "all") params.set("type", type);
      if (account !== "all") params.set("account_id", account);
      if (jar !== "all") params.set("jar_id", jar);
      if (category !== "all") {
        params.set("category", category);
      }

      if (search.trim()) {
        params.set("search", search.trim());
      }

      if (dateRange) {
        params.set("from_date", dateRange.start);
        params.set("to_date", dateRange.end);
      }

      const data = await apiFetch<TransactionsResponse>(
        `/api/v1/transactions?${params.toString()}`,
      );

      setTransactions(data.transactions);
      setTotal(data.total);
    } catch (err) {
      console.error(err);
      setError("Unable to load transactions.");
    } finally {
      setLoading(false);
    }
  }, [
    limit,
    offset,
    type,
    account,
    jar,
    category,
    search,
    dateRange,
  ]);

  useEffect(() => {
    fetchTransactions();
  }, [fetchTransactions]);

  function resetFilters() {
    setSearch("");
    setType("all");
    setAccount("all");
    setJar("all");
    setCategory("all");
    setDateRange(null);
    setPage(1);
  }

  const groups = useMemo(
    () => groupTransactions(transactions),
    [transactions],
  );

  const hasActiveFilters =
    type !== "all" ||
    account !== "all" ||
    jar !== "all" ||
    category !== "all" ||
    dateRange !== null ||
    search.trim() !== "";

  return (
    <div className="flex w-full flex-col gap-5">
      <TransactionsFilters
        search={search}
        type={type}
        account={account}
        jar={jar}
        category={category}
        dateRange={dateRange}
        accounts={accounts}
        jars={jars}
        filterOpen={filterOpen}
        loadingFilters={loadingFilters}
        hasActiveFilters={hasActiveFilters}
        onSearchChange={(value) => {
          setSearch(value);
          setPage(1);
        }}
        onTypeChange={(value) => {
          setType(value);
          setPage(1);
        }}
        onAccountChange={(value) => {
          setAccount(value);
          setPage(1);
        }}
        onJarChange={(value) => {
          setJar(value);
          setPage(1);
        }}
        onCategoryChange={(value) => {
          setCategory(value);
          setPage(1);
        }}
        onDateRangeChange={(value) => {
          setDateRange(value);
          setPage(1);
        }}
        onFilterToggle={() =>
          setFilterOpen((current) => !current)
        }
        onReset={resetFilters}
      />

      {loading && <TransactionsSkeleton />}

      {!loading && error && (
        <div>
          {/* error component */}
        </div>
      )}

      {!loading && !error && transactions.length === 0 && (
        <TransactionsEmpty onClear={resetFilters} />
      )}

      {!loading && !error && transactions.length > 0 && (
        <div className="flex flex-col gap-6">
          {groups.map(([group, items]) => (
            <TransactionGroup
              key={group}
              group={group}
              transactions={items}
            />
          ))}

          <TransactionsPagination
            page={page}
            totalPages={totalPages}
            limit={limit}
            offset={offset}
            transactionCount={transactions.length}
            total={total}
            onPageChange={setPage}
            onLimitChange={(value) => {
              setLimit(value);
              setPage(1);
            }}
          />
        </div>
      )}
    </div>
  );
}
