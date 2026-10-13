"use client";

import { Suspense, useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import {
  ArrowBack, ArrowDownward, ArrowOutward, Check, CheckCircle,
  Delete, Schedule, CalendarToday,
  ShoppingCart, Restaurant,
  DirectionsCar, Movie, Medication, MoreHoriz,
  TrendingUp, VolunteerActivism, Home, SwapHoriz,
} from "@material-symbols-svg/react/w400";
import { apiFetch } from "@/lib/api";
import type { Account } from "@/lib/accounts/accounts";
import AccountIcon from "@/components/accounts/AccountIcon";
import { getJarIcon } from "@/components/jars/jar-icons";
import { formatMoney, formatRFC3339, type TransactionCategory, type TransactionType } from "@/components/transactions/transaction-utils";
import DeleteTransactionDialog from "@/components/transactions/DeleteTransactionDialog";

type TransactionDetails = {
  id: string;
  name: string;
  type: TransactionType;
  category: TransactionCategory;
  from_account_id: string | null;
  to_account_id: string | null;
  jar_id: string | null;
  amount: number;
  occurred_at: string;
  is_master_income: boolean;
  updated_at: string;
};
type Jar = { id: string; name: string; is_archived?: boolean; icon_key?: string | null; available_amount?: number };
type EditForm = {
  name: string; type: TransactionType; category: TransactionCategory; accountId: string;
  toAccountId: string; jarId: string; amount: string;
  date: string; time: string; isMasterIncome: boolean;
};

const categories: { value: TransactionCategory; label: string; icon: typeof ShoppingCart }[] = [
  { value: "groceries", label: "Groceries", icon: ShoppingCart },
  { value: "food", label: "Food", icon: Restaurant },
  { value: "transport", label: "Transport", icon: DirectionsCar },
  { value: "entertainment", label: "Fun", icon: Movie },
  { value: "health", label: "Health", icon: Medication },
  { value: "investment", label: "Investing", icon: TrendingUp },
  { value: "housing", label: "Housing", icon: Home },
  { value: "donation", label: "Giving", icon: VolunteerActivism },
  { value: "transfer", label: "Transfer", icon: SwapHoriz },
  { value: "other", label: "Other", icon: MoreHoriz },
];
function localParts(value: string) {
  const date = new Date(value);
  if (!Number.isFinite(date.getTime())) return { date: "", time: "" };
  const pad = (v: number) => String(v).padStart(2, "0");
  return { date: `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`, time: `${pad(date.getHours())}:${pad(date.getMinutes())}` };
}

export default function EditTransactionPage() {
  return <Suspense fallback={<div className="min-h-screen bg-[#f8f9ff] p-6 pt-24 text-sm text-slate-500">Loading transaction…</div>}><EditTransactionContent /></Suspense>;
}

function EditTransactionContent() {
  const searchParams = useSearchParams();
  const id = searchParams.get("id") ?? "";
  const router = useRouter();
  const [transaction, setTransaction] = useState<TransactionDetails | null>(null);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [jars, setJars] = useState<Jar[]>([]);
  const [form, setForm] = useState<EditForm | null>(null);
  const [loading, setLoading] = useState(Boolean(id));
  const [optionsLoading, setOptionsLoading] = useState(Boolean(id));
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(id ? "" : "Transaction ID is missing.");
  const [toast, setToast] = useState("");
  const [pendingDelete, setPendingDelete] = useState<TransactionDetails | null>(null);

  useEffect(() => {
    let active = true;
    if (!id) return;
    Promise.all([
      apiFetch<Account[]>("/api/v1/accounts"),
      apiFetch<Jar[]>("/api/v1/jars"),
    ]).then(([accountList, jarList]) => {
      if (!active) return;
      setAccounts(
        accountList
          .filter((account) => !account.is_archived)
          .sort((a, b) => Number(b.is_primary) - Number(a.is_primary))
      );
      setJars(jarList.filter((jar) => !jar.is_archived));
    }).catch((cause: unknown) => {
      if (active) setError(cause instanceof Error ? cause.message : "Unable to load account and jar options.");
    }).finally(() => { if (active) setOptionsLoading(false); });

    apiFetch<TransactionDetails>(`/api/v1/transactions/${id}`).then((txn) => {
      if (!active) return;
      setTransaction(txn);
      const parts = localParts(txn.occurred_at);
      setForm({
        name: txn.name, type: txn.type, category: txn.category, amount: (txn.amount / 100).toFixed(2),
        accountId: txn.from_account_id ?? txn.to_account_id ?? "", toAccountId: txn.to_account_id ?? "",
        jarId: txn.jar_id ?? "", date: parts.date, time: parts.time,
        isMasterIncome: txn.is_master_income,
      });
      setLoading(false);
    }).catch((cause: unknown) => {
      if (active) setError(cause instanceof Error ? cause.message : "Unable to load this transaction.");
    }).finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [id]);

  const setField = useCallback(<K extends keyof EditForm>(key: K, value: EditForm[K]) => {
    setForm((current) => current ? { ...current, [key]: value } : current);
  }, []);
  const typeTheme = useMemo(() => {
    if (form?.type === "income") {
      return {
        accent: "text-emerald-500",
        active: "bg-emerald-500 text-white",
        selected: "border-2 border-emerald-500 bg-gradient-to-br from-emerald-50/70 via-white to-teal-50/40 shadow-sm shadow-emerald-500/10",
        check: "bg-emerald-500 text-white shadow-sm shadow-emerald-500/30",
        button: "bg-emerald-500 hover:bg-emerald-600 shadow-emerald-500/25",
        jar: "border-2 border-emerald-500 bg-gradient-to-br from-emerald-50/50 via-white to-teal-50/30 shadow-sm shadow-emerald-500/10",
        categorySelected: "border-2 border-emerald-500 bg-gradient-to-r from-emerald-50 via-teal-50/50 to-white text-emerald-600 font-bold shadow-sm shadow-emerald-500/10",
        categoryIconSelected: "bg-emerald-500 text-white shadow-sm shadow-emerald-500/20",
        categoryCheck: "text-emerald-500",
      };
    }
    if (form?.type === "transfer") {
      return {
        accent: "text-blue-600",
        active: "bg-blue-600 text-white",
        selected: "border-2 border-blue-500 bg-gradient-to-br from-blue-50/70 via-white to-indigo-50/40 shadow-sm shadow-blue-500/10",
        check: "bg-blue-600 text-white shadow-sm shadow-blue-500/30",
        button: "bg-blue-600 hover:bg-blue-700 shadow-blue-500/25",
        jar: "border-2 border-blue-500 bg-gradient-to-br from-blue-50/50 via-white to-indigo-50/30 shadow-sm shadow-blue-500/10",
        categorySelected: "border-2 border-blue-500 bg-gradient-to-r from-blue-50 via-indigo-50/50 to-white text-blue-600 font-bold shadow-sm shadow-blue-500/10",
        categoryIconSelected: "bg-blue-600 text-white shadow-sm shadow-blue-500/20",
        categoryCheck: "text-blue-600",
      };
    }
    return {
      accent: "text-red-500",
      active: "bg-red-500 text-white",
      selected: "border-2 border-red-500 bg-gradient-to-br from-red-50/70 via-white to-rose-50/40 shadow-sm shadow-red-500/10",
      check: "bg-red-500 text-white shadow-sm shadow-red-500/30",
      button: "bg-red-500 hover:bg-red-600 shadow-red-500/25",
      jar: "border-2 border-red-500 bg-gradient-to-br from-red-50/50 via-white to-rose-50/30 shadow-sm shadow-red-500/10",
      categorySelected: "border-2 border-red-500 bg-gradient-to-r from-red-50 via-rose-50/50 to-white text-red-600 font-bold shadow-sm shadow-red-500/10",
      categoryIconSelected: "bg-red-500 text-white shadow-sm shadow-red-500/20",
      categoryCheck: "text-red-500",
    };
  }, [form?.type]);

  function switchType(type: TransactionType) {
    setForm((current) => {
      if (!current) return current;
      const accountId = type === "income" && current.type === "transfer"
        ? current.toAccountId || current.accountId
        : current.accountId;
      const toAccountId = type === "transfer" && current.toAccountId === accountId ? "" : current.toAccountId;
      return { ...current, type, accountId, toAccountId };
    });
  }

  async function save() {
    if (!transaction || !form) return;
    if (form.type !== transaction.type) {
      setError("This transaction’s type cannot be changed with the current update API. Switch back to its original type to save.");
      return;
    }
    const amount = Number(form.amount);
    if (!form.name.trim() || !Number.isFinite(amount) || amount <= 0 || !form.date || !form.time) {
      setError("Enter a title, a positive amount, and a valid date and time."); return;
    }
    const [year, month, day] = form.date.split("-").map(Number);
    const [hours, minutes] = form.time.split(":").map(Number);
    const occurredAt = new Date(year, month - 1, day, hours, minutes, 0, 0);
    setSaving(true); setError("");
    try {
      await apiFetch(`/api/v1/transactions/${id}`, {
        method: "PATCH",
        body: JSON.stringify({
          name: form.name.trim(), category: form.category,
          from_account_id: form.type === "income" ? "" : form.accountId,
          to_account_id: form.type === "expense" ? "" : form.type === "transfer" ? form.toAccountId : form.accountId,
          jar_id: form.jarId, amount: Math.round(amount * 100),
          occurred_at: formatRFC3339(occurredAt), is_master_income: form.type === "income" && form.isMasterIncome,
        }),
      });
      setToast("Transaction updated successfully");
      window.setTimeout(() => router.push("/transactions"), 650);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Could not save transaction.");
    } finally { setSaving(false); }
  }

  if (loading) return <div className="min-h-screen bg-[#f8f9ff] p-6 pt-24 text-sm text-slate-500">Loading transaction…</div>;
  if (!transaction || !form) return <div className="min-h-screen bg-[#f8f9ff] p-6 pt-24"><p className="text-sm text-rose-700">{error || "Transaction not found."}</p><Link href="/transactions" className="mt-4 inline-block text-sm font-semibold text-blue-700">Back to transactions</Link></div>;

  const shownJars = jars;
  const jarColors = ["#F43F5E", "#6366F1", "#07B682", "#F6A723"];

  return <div className="min-h-screen bg-[#f8f9ff] pb-28 text-[#0b1c30]">
    <header className="fixed inset-x-0 top-0 z-30 border-b border-slate-200/50 bg-[#f8f9ff]/90 pt-safe backdrop-blur-xl">
      <div className="mx-auto flex h-16 max-w-xl items-center justify-between px-4">
        <button aria-label="Go back" onClick={() => router.back()} className="flex h-11 w-11 items-center justify-center rounded-full hover:bg-slate-100 text-slate-700"><ArrowBack size={24} /></button>
        <h1 className="text-base font-semibold">Edit Transaction</h1>
        <div className="w-11" />
      </div>
    </header>
    <main className="mx-auto flex max-w-xl flex-col gap-4 px-4 pt-20">
      <div className="grid grid-cols-3 rounded-full bg-slate-100 p-1.5 shadow-sm">
        {(["expense", "income", "transfer"] as TransactionType[]).map((type) => (
          <button
            type="button"
            key={type}
            aria-pressed={form.type === type}
            onClick={() => switchType(type)}
            className={`flex items-center justify-center gap-1 rounded-full py-2.5 text-sm font-semibold capitalize transition-colors ${form.type === type ? typeTheme.active : "text-slate-500"}`}
          >
            {type === "expense" ? <ArrowOutward size={17} /> : type === "income" ? <ArrowDownward size={17} /> : <SwapHoriz size={17} />} {type}
          </button>
        ))}
      </div>
      <section className="space-y-3 rounded-3xl bg-white p-5 shadow-sm">
        <div className="flex items-center gap-2 pb-2 border-b border-slate-100">
          <input
            id="transaction-name"
            maxLength={120}
            value={form.name}
            onChange={(e) => setField("name", e.target.value)}
            className="min-w-0 flex-1 bg-transparent text-sm font-medium text-slate-700 placeholder:text-slate-400 outline-none"
            placeholder="What was this for?"
          />
        </div>
        <span className="block text-xs font-bold uppercase tracking-wider text-slate-400">Transaction Amount</span>
        <div className="flex items-center gap-1">
          <span className={`text-3xl font-extrabold ${typeTheme.accent}`}>
            {form.type === "expense" ? "−₹" : form.type === "income" ? "+₹" : "⇄₹"}
          </span>
          <input
            aria-label="Transaction amount"
            inputMode="decimal"
            type="number"
            min="0.01"
            step="0.01"
            value={form.amount}
            onChange={(e) => setField("amount", e.target.value)}
            className="min-w-0 flex-1 bg-transparent text-3xl font-extrabold text-slate-900 outline-none"
          />
        </div>
        <div className="flex items-center justify-between gap-2 pt-1">
          <div className="flex flex-wrap gap-2">
            {[-10, 10, 100].map((delta) => (
              <button
                key={delta}
                type="button"
                onClick={() => {
                  const currentVal = parseFloat(form.amount) || 0;
                  setField("amount", Math.max(0, currentVal + delta).toFixed(2));
                }}
                className="rounded-full bg-slate-100 hover:bg-slate-200 px-3 py-1.5 text-xs font-semibold text-slate-700 transition"
              >
                {delta < 0 ? "−" : "+"}₹{Math.abs(delta)}
              </button>
            ))}
          </div>
          <span className="text-xs font-semibold text-slate-400">
            {formatMoney(transaction.amount)}
          </span>
        </div>
      </section>
      {form.type === "income" && (
        <button
          type="button"
          role="switch"
          aria-checked={form.isMasterIncome}
          onClick={() => setField("isMasterIncome", !form.isMasterIncome)}
          className="flex items-center justify-between rounded-2xl bg-white px-4 py-3.5 text-left shadow-sm"
        >
          <span>
            <span className="block text-sm font-semibold">Master income</span>
            <span className="mt-0.5 block text-xs text-slate-500">Count this as recurring or primary income</span>
          </span>
          <span className={`relative h-6 w-11 rounded-full transition-colors ${form.isMasterIncome ? "bg-emerald-500" : "bg-slate-300"}`}>
            <span className={`absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition-transform ${form.isMasterIncome ? "translate-x-5" : "translate-x-0.5"}`} />
          </span>
        </button>
      )}
      <section className="space-y-3 rounded-2xl bg-white p-4 shadow-sm">
        <div className="flex items-center justify-between">
          <span className="text-xs font-bold uppercase tracking-wider text-slate-500">
            {form.type === "income" ? "Receive To Account" : form.type === "transfer" ? "From Account" : "Pay From Account"}
          </span>
          <Link href="/accounts" className="text-xs font-semibold text-slate-500 hover:text-slate-800">Manage</Link>
        </div>
        {optionsLoading && <p className="text-sm text-slate-500">Loading accounts…</p>}
        <div className="grid grid-cols-2 gap-2.5">
          {accounts.map((account, index) => {
            const selected = form.accountId === account.id;
            return (
              <button
                type="button"
                key={account.id}
                onClick={() => setField("accountId", account.id)}
                className={`relative flex min-h-24 flex-col justify-between rounded-2xl border p-3 text-left transition active:scale-[.98] ${selected ? `${typeTheme.selected}` : "border-slate-200 bg-white hover:border-slate-300"
                  }`}
              >
                <div className="mb-2 flex items-start justify-between">
                  <AccountIcon iconKey={account.icon_key ?? ""} name={account.name} index={index} size="sm" />
                  <span className={`flex h-5 w-5 items-center justify-center rounded-full ${selected ? `${typeTheme.check}` : "border border-slate-300"}`}>
                    {selected && <Check size={13} />}
                  </span>
                </div>
                <span className="truncate text-xs font-bold">{account.name}</span>
                {account.is_primary && (
                  <span className="mt-1 w-fit rounded bg-slate-100 px-1.5 py-0.5 text-[9px] font-bold uppercase text-slate-500">
                    Primary
                  </span>
                )}
                <span className="mt-1 text-sm font-extrabold">{formatMoney(account.balance)}</span>
              </button>
            );
          })}
        </div>
      </section>
      {form.type === "transfer" && (
        <section className="space-y-3 rounded-2xl bg-white p-4 shadow-sm">
          <span className="text-xs font-bold uppercase tracking-wider text-slate-500">To Account</span>
          <div className="grid grid-cols-2 gap-2.5">
            {accounts.filter((account) => account.id !== form.accountId).map((account, index) => {
              const selected = form.toAccountId === account.id;
              return (
                <button
                  type="button"
                  key={account.id}
                  onClick={() => setField("toAccountId", account.id)}
                  className={`relative flex min-h-24 flex-col justify-between rounded-2xl border p-3 text-left transition active:scale-[.98] ${selected ? "border-2 border-blue-500 bg-gradient-to-br from-blue-50/70 via-white to-indigo-50/40 shadow-md shadow-blue-500/10" : "border-slate-200 bg-white hover:border-slate-300"
                    }`}
                >
                  <div className="mb-2 flex items-start justify-between">
                    <AccountIcon iconKey={account.icon_key ?? ""} name={account.name} index={index} size="sm" />
                    <span className={`flex h-5 w-5 items-center justify-center rounded-full ${selected ? "bg-blue-600 text-white shadow-sm shadow-blue-500/30" : "border border-slate-300"}`}>
                      {selected && <Check size={13} />}
                    </span>
                  </div>
                  <span className="truncate text-xs font-bold">{account.name}</span>
                  {account.is_primary && (
                    <span className="mt-1 w-fit rounded bg-slate-100 px-1.5 py-0.5 text-[9px] font-bold uppercase text-slate-500">
                      Primary
                    </span>
                  )}
                  <span className="mt-1 text-sm font-extrabold">{formatMoney(account.balance)}</span>
                </button>
              );
            })}
          </div>
        </section>
      )}
      <section className="space-y-3 rounded-2xl bg-white p-4 shadow-sm">
        <div className="flex items-center justify-between">
          <span className="text-xs font-bold uppercase tracking-wider text-slate-500">Jar</span>
          {form.jarId && (
            <button type="button" onClick={() => setField("jarId", "")} className="text-xs font-semibold text-slate-500 hover:text-slate-800">
              Clear jar
            </button>
          )}
        </div>
        {optionsLoading && <p className="text-sm text-slate-500">Loading jars…</p>}
        <div className="grid grid-cols-2 gap-2.5">
          {shownJars.map((jar, i) => {
            const Icon = getJarIcon(jar.icon_key ?? jar.name);
            const selected = form.jarId === jar.id;
            const jarColor = jarColors[i % jarColors.length];
            return (
              <button
                type="button"
                key={jar.id}
                onClick={() => setField("jarId", selected ? "" : jar.id)}
                className={`relative flex min-h-24 flex-col justify-between rounded-2xl border p-3.5 text-left transition active:scale-[.98] ${selected ? typeTheme.jar : "border-slate-200 bg-white hover:border-slate-300"
                  }`}
              >
                <div className="flex items-start justify-between">
                  <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl" style={{ backgroundColor: `${jarColor}18` }}>
                    <Icon size={18} color={jarColor} />
                  </span>
                  <span className={`flex h-5 w-5 items-center justify-center rounded-full ${selected ? `${typeTheme.check}` : "border border-slate-300"}`}>
                    {selected && <Check size={13} />}
                  </span>
                </div>
                <div>
                  <span className="block truncate text-xs font-bold text-slate-900">{jar.name}</span>
                  <span className="mt-1 block text-sm font-extrabold text-slate-900 font-mono">{formatMoney(jar.available_amount ?? 0)}</span>
                </div>
              </button>
            );
          })}
        </div>
      </section>
      <section className="space-y-3 rounded-2xl bg-white p-4 shadow-sm">
        <div className="flex items-center justify-between">
          <span className="text-xs font-bold uppercase tracking-wider text-slate-500">Category</span>
          <span className="text-[11px] font-bold text-slate-500 bg-slate-100 px-2 py-0.5 rounded-full">
            {categories.find((c) => c.value === form.category)?.label || "Selected"}
          </span>
        </div>
        <div className="grid grid-cols-2 gap-2">
          {categories.map(({ value, label, icon: Icon }) => {
            const isSelected = form.category === value;
            return (
              <button
                type="button"
                key={value}
                onClick={() => setField("category", value)}
                className={`flex items-center gap-2.5 p-2.5 rounded-xl text-left transition-all active:scale-95 border ${isSelected
                  ? typeTheme.categorySelected
                  : "border-slate-200 bg-white hover:border-slate-300 hover:bg-slate-50 text-slate-700"
                  }`}
              >
                <div className={`w-8 h-8 rounded-lg flex items-center justify-center shrink-0 transition-transform ${isSelected ? typeTheme.categoryIconSelected : "bg-slate-100 text-slate-600"
                  }`}>
                  <Icon size={18} />
                </div>
                <span className="font-bold text-xs truncate flex-1">{label}</span>
                {isSelected && <CheckCircle size={16} className={`ml-auto shrink-0 ${typeTheme.categoryCheck}`} />}
              </button>
            );
          })}
        </div>
      </section>
      <section className="grid grid-cols-2 gap-3">
        <label className="flex items-center gap-2 rounded-2xl bg-white p-3 shadow-sm">
          <CalendarToday className="shrink-0 text-slate-500" size={20} />
          <span className="min-w-0 flex-1">
            <span className="block text-xs text-slate-500">Date</span>
            <input aria-label="Date" type="date" value={form.date} onChange={(e) => setField("date", e.target.value)} className="w-full bg-transparent text-sm font-bold outline-none" />
          </span>
        </label>
        <label className="flex items-center gap-2 rounded-2xl bg-white p-3 shadow-sm">
          <Schedule className="shrink-0 text-slate-500" size={20} />
          <span className="min-w-0 flex-1">
            <span className="block text-xs text-slate-500">Time</span>
            <input aria-label="Time" type="time" value={form.time} onChange={(e) => setField("time", e.target.value)} className="w-full bg-transparent text-sm font-bold outline-none" />
          </span>
        </label>
      </section>
      {error && <p role="alert" className="rounded-xl bg-rose-50 px-3 py-2 text-sm text-rose-700">{error}</p>}
      <div className="flex flex-col gap-3 pt-1">
        <button disabled={saving || optionsLoading} onClick={save} className={`flex h-[52px] items-center justify-center rounded-full text-sm font-bold text-white shadow-md disabled:opacity-60 transition ${typeTheme.button}`}>
          <Check size={20} className="mr-2" />
          {saving ? "Saving…" : optionsLoading ? "Loading options…" : "Save Changes"}
        </button>
        <button disabled={saving} onClick={() => router.back()} className="h-11 rounded-full bg-slate-100 hover:bg-slate-200 text-sm font-semibold text-slate-700 transition">
          Discard Changes
        </button>
      </div>
    </main>
    {toast && <div role="status" className="fixed inset-x-4 bottom-24 z-40 mx-auto flex max-w-lg items-center gap-3 rounded-2xl bg-[#213145] p-4 text-sm font-semibold text-white shadow-xl"><CheckCircle className="text-emerald-300" size={22} />{toast}</div>}
  </div>;
}
