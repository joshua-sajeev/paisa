"use client";

import { useState } from "react";
import { 
  ArrowBack, 
  Close, 
  CheckCircle, 
  Restaurant, 
  CalendarToday, 
  Schedule,
  EditNote
} from "@material-symbols-svg/react/w400";
import { TransactionFormData } from "@/app/(protected)/transactions/add/page";
import { apiFetch } from "@/lib/api";
import AccountIcon from "@/components/accounts/AccountIcon";
import { getJarIcon } from "@/components/jars/jar-icons";
import { 
  formatMoney, 
  CATEGORIES, 
  formatTime, 
  formatRFC3339, 
  formatTimeInput 
} from "@/components/transactions/transaction-utils";

function getValidDate(value: string): Date {
  const date = new Date(value);
  return Number.isFinite(date.getTime()) ? date : new Date();
}

type Step3Props = {
  formData: TransactionFormData;
  updateFormData: (data: Partial<TransactionFormData>) => void;
  onBack: () => void;
  onCancel?: () => void;
  onComplete: (addAnother: boolean, occurredAt: string) => void;
};

export default function Step3Confirmation({ formData, updateFormData, onBack, onCancel, onComplete }: Step3Props) {
  const [loading, setLoading] = useState(false);
  const [nameError, setNameError] = useState<string | null>(null);
  const [addAnother, setAddAnother] = useState(false);
  const occurredAtDate = getValidDate(formData.occurredAt);
  const safeOccurredAt = Number.isFinite(new Date(formData.occurredAt).getTime())
    ? formData.occurredAt
    : formatRFC3339(occurredAtDate);
  const getInitialDateMode = () => {
    const d = getValidDate(formData.occurredAt);
    const today = new Date();
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);

    if (d.toDateString() === today.toDateString()) return 'today';
    if (d.toDateString() === yesterday.toDateString()) return 'yesterday';
    return 'custom';
  };

  const [dateMode, setDateMode] = useState<'today' | 'yesterday' | 'custom'>(getInitialDateMode());

  const handleSave = async () => {
    const trimmedName = formData.name.trim();
    if (!trimmedName) {
      setNameError("Transaction name is required");
      return;
    }

    setLoading(true);
    setNameError(null);
    try {
      const occurredAtRFC3339 = formatRFC3339(occurredAtDate);
      if (safeOccurredAt !== formData.occurredAt) {
        updateFormData({ occurredAt: safeOccurredAt });
      }
      const payload = {
        name: trimmedName,
        type: formData.type,
        category: formData.category,
        from_account_id: formData.type === 'income' ? "" : formData.fromAccountId,
        to_account_id: formData.type === 'expense' ? "" : formData.toAccountId,
        jar_id: formData.jarId || "",
        amount: Math.round(parseFloat(formData.amount) * 100),
        occurred_at: occurredAtRFC3339,
        is_master_income: formData.isMasterIncome
      };

      await apiFetch("/api/v1/transactions/", {
        method: "POST",
        body: JSON.stringify(payload)
      });

      onComplete(addAnother, occurredAtRFC3339);
    } catch (err) {
      console.error("Failed to save transaction", err);
      alert("Failed to save transaction. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  const setOccurredAt = (mode: 'today' | 'yesterday' | 'custom', customDate?: string) => {
    setDateMode(mode);
    const date = getValidDate(formData.occurredAt);
    const now = new Date();
    if (mode === 'today') {
      date.setFullYear(now.getFullYear(), now.getMonth(), now.getDate());
    } else if (mode === 'yesterday') {
      now.setDate(now.getDate() - 1);
      date.setFullYear(now.getFullYear(), now.getMonth(), now.getDate());
    } else if (mode === 'custom' && customDate) {
      const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(customDate);
      if (!match) return;
      const [, year, month, day] = match;
      const y = Number(year);
      const m = Number(month);
      const d = Number(day);
      const selectedDate = new Date(y, m - 1, d);
      if (selectedDate.getFullYear() !== y || selectedDate.getMonth() !== m - 1 || selectedDate.getDate() !== d) return;
      date.setFullYear(y, m - 1, d);
    }
    updateFormData({ occurredAt: formatRFC3339(date) });
  };

  const handleTimeChange = (timeString: string) => {
    if (!timeString) return;
    const [h, m] = timeString.split(":").map(Number);
    if (isNaN(h) || isNaN(m)) return;
    const date = getValidDate(formData.occurredAt);
    date.setHours(h, m, 0, 0);
    updateFormData({ occurredAt: formatRFC3339(date) });
  };

  const handleSetCurrentTime = () => {
    const now = new Date();
    const date = getValidDate(formData.occurredAt);
    date.setHours(now.getHours(), now.getMinutes(), now.getSeconds(), 0);
    updateFormData({ occurredAt: formatRFC3339(date) });
  };

  const categoryLabel = CATEGORIES.find(c => c.value === formData.category)?.label || 'Other';
  const currentAmount = parseFloat(formData.amount) * 100;
  const transactionAmount = Number.isFinite(currentAmount) ? Math.round(currentAmount) : 0;
  const accountImpacts = formData.type === 'income'
    ? [{
        name: formData.toAccountName,
        iconKey: formData.toAccountIconKey,
        detail: 'Deposit to',
        current: formData.toAccountBalance,
        future: formData.toAccountBalance + transactionAmount,
      }]
    : formData.type === 'expense'
      ? [{
          name: formData.fromAccountName,
          iconKey: formData.fromAccountIconKey,
          detail: 'Payment from',
          current: formData.fromAccountBalance,
          future: formData.fromAccountBalance - transactionAmount,
        }]
      : [
          {
            name: formData.fromAccountName,
            iconKey: formData.fromAccountIconKey,
            detail: 'Transfer from',
            current: formData.fromAccountBalance,
            future: formData.fromAccountBalance - transactionAmount,
          },
          {
            name: formData.toAccountName,
            iconKey: formData.toAccountIconKey,
            detail: 'Transfer to',
            current: formData.toAccountBalance,
            future: formData.toAccountBalance + transactionAmount,
          },
        ];
  const jarChange = formData.type === 'expense'
    ? -transactionAmount
    : formData.type === 'income' && !formData.isMasterIncome
      ? transactionAmount
      : 0;
  const futureJarBalance = formData.jarBalance + jarChange;
  const jarNameLower = formData.jarName.toLowerCase();
  const jarIconKey = jarNameLower.includes("necessit")
    ? "necessities"
    : jarNameLower.includes("leisure")
      ? "leisure"
      : jarNameLower.includes("invest")
        ? "investment"
        : jarNameLower.includes("giv")
          ? "giving"
          : formData.jarIconKey || "jar";
  const JarIcon = getJarIcon(jarIconKey);

  const typeTheme = {
    expense: {
      stepText: 'text-red-500',
      pillBg: 'bg-red-500',
      btnBg: 'bg-red-500 hover:bg-red-600 shadow-red-500/25',
      focusRing: 'focus-within:border-red-500 focus-within:ring-2 focus-within:ring-red-500/10',
    },
    income: {
      stepText: 'text-emerald-500',
      pillBg: 'bg-emerald-500',
      btnBg: 'bg-emerald-500 hover:bg-emerald-600 shadow-emerald-500/25',
      focusRing: 'focus-within:border-emerald-500 focus-within:ring-2 focus-within:ring-emerald-500/10',
    },
    transfer: {
      stepText: 'text-blue-600',
      pillBg: 'bg-blue-600',
      btnBg: 'bg-blue-600 hover:bg-blue-700 shadow-blue-500/25',
      focusRing: 'focus-within:border-blue-600 focus-within:ring-2 focus-within:ring-blue-600/10',
    },
  }[formData.type];

  return (
    <div className="flex flex-col flex-1 w-full bg-white min-h-screen relative overflow-y-auto no-scrollbar p-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <button onClick={onBack} className="w-10 h-10 flex items-center justify-center rounded-full bg-slate-100 text-slate-900 active:scale-95 transition-transform">
          <ArrowBack size={20} />
        </button>
        <div className="flex flex-col items-center">
          <span className={`text-xs font-bold ${typeTheme.stepText} tracking-widest uppercase`}>Step 3 of 3</span>
        </div>
        <button 
          onClick={onCancel || onBack}
          className="w-10 h-10 flex items-center justify-center rounded-full bg-slate-100 text-slate-500 hover:text-slate-800 active:scale-95 transition-colors"
        >
          <Close size={20} />
        </button>
      </div>

      {/* Hero Card */}
      <section className="w-full bg-white rounded-xl p-6 shadow-sm flex flex-col items-center text-center relative overflow-hidden mb-6 border border-slate-100">
        <div className="flex items-center gap-2 mb-2">
          <span className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full ${typeTheme.pillBg} text-white text-[10px] font-bold tracking-wider uppercase`}>
            <span className="w-1.5 h-1.5 rounded-full bg-white"></span>
            {formData.type}
          </span>
        </div>
        <div className="flex items-baseline justify-center gap-0.5 my-1">
          <span className="text-4xl font-extrabold tracking-tight text-slate-900">{formatMoney(currentAmount).split('.')[0]}</span>
          <span className="text-xl text-slate-400 font-bold">.{formatMoney(currentAmount).split('.')[1]?.replace('₹', '') || '00'}</span>
        </div>
        <div className="mt-3 flex w-full max-w-md items-center gap-2 rounded-xl bg-blue-50/90 px-3.5 py-2 text-left transition-all focus-within:ring-2 focus-within:ring-blue-500/15">
          <input
            id="tx-name-input"
            type="text"
            aria-label="Transaction name"
            value={formData.name}
            onChange={(e) => {
              updateFormData({ name: e.target.value });
              if (nameError) setNameError(null);
            }}
            placeholder={
              formData.type === 'expense'
                ? "Name this expense..."
                : formData.type === 'income'
                ? "Name this income..."
                : "Name this transfer..."
            }
            className="min-w-0 flex-1 border-0 bg-transparent p-0 text-center text-base font-semibold tracking-tight text-slate-900 placeholder:text-slate-500 focus:outline-none focus:ring-0"
          />
          <EditNote size={18} className="shrink-0 text-slate-600" />
        </div>
        {nameError && <p className="mt-1 text-xs font-medium text-red-600">{nameError}</p>}
      </section>

      {/* Details */}
      <section className="w-full bg-white rounded-xl p-4 shadow-sm mb-6 border border-slate-100">
        <div className="flex items-center justify-between pb-2 mb-2 border-b border-slate-50">
          <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Ledger Details</span>
        </div>
        
        {/* Account balance after transaction */}
        {accountImpacts.map((impact, index) => (
          <div key={`${impact.detail}-${index}`} className="flex items-center justify-between gap-3 border-b border-slate-50 py-3 last:border-b-0">
            <div className="flex min-w-0 items-center gap-3">
              <AccountIcon iconKey={impact.iconKey} name={impact.name} size="sm" />
              <div className="flex min-w-0 flex-col">
                <span className="text-[10px] font-bold uppercase text-slate-400">{impact.detail}</span>
                <span className="truncate text-sm font-bold text-slate-900">{impact.name || 'Account'}</span>
                <span className="text-[11px] font-medium text-slate-500">Now {formatMoney(impact.current)}</span>
              </div>
            </div>
            <div className="shrink-0 text-right">
              <span className="block text-[10px] font-bold uppercase tracking-wide text-slate-400">After txn</span>
              <span className="text-sm font-extrabold tabular-nums text-slate-900">{formatMoney(impact.future)}</span>
            </div>
          </div>
        ))}

        {/* Jar */}
        {formData.jarId && (
          <div className="flex items-center justify-between gap-3 py-3">
            <div className="flex items-center gap-3 min-w-0">
              <div className="w-10 h-10 rounded-full bg-slate-100 flex items-center justify-center shrink-0 text-purple-600">
                <JarIcon size={20} color="currentColor" />
              </div>
              <div className="flex min-w-0 flex-col">
                <span className="text-[10px] font-bold text-slate-400 uppercase">Vault / Jar</span>
                <span className="text-sm font-bold text-slate-900 truncate">{formData.jarName}</span>
                <span className="text-[11px] font-medium text-slate-500">Now {formatMoney(formData.jarBalance)}</span>
              </div>
            </div>
            <div className="shrink-0 text-right">
              <span className="block text-[10px] font-bold uppercase tracking-wide text-slate-400">After txn</span>
              {formData.isMasterIncome ? (
                <span className="text-xs font-bold text-slate-500">Auto allocation</span>
              ) : (
                <span className="text-sm font-extrabold tabular-nums text-slate-900">{formatMoney(futureJarBalance)}</span>
              )}
            </div>
          </div>
        )}

        {/* Category */}
        <div className="flex items-center justify-between border-t border-slate-50 py-3">
          <div className="flex items-center gap-3 min-w-0">
            <div className="w-10 h-10 rounded-full bg-slate-100 flex items-center justify-center shrink-0 text-emerald-600">
              <Restaurant size={20} />
            </div>
            <div className="flex flex-col min-w-0">
              <span className="text-[10px] font-bold text-slate-400 uppercase">Category</span>
              <span className="text-sm font-bold text-slate-900 truncate">{categoryLabel}</span>
            </div>
          </div>
        </div>
      </section>

      {/* Date Allocation */}
      <section className="w-full bg-white rounded-xl p-4 shadow-sm mb-8 border border-slate-100">
        <div className="flex items-center justify-between mb-2">
          <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Posting Date</span>
          <span className="text-xs font-bold text-blue-600">
            {dateMode === 'today' ? 'Today' : dateMode === 'yesterday' ? 'Yesterday' : occurredAtDate.toLocaleDateString()}
          </span>
        </div>
        <div className="grid grid-cols-3 gap-2 p-1 bg-slate-50 rounded-xl">
          <button 
            onClick={() => setOccurredAt('today')}
            className={`py-2 px-2 rounded-lg text-center text-xs font-bold transition-all ${dateMode === 'today' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-500 hover:text-slate-900'}`}
          >
            Today
          </button>
          <button 
            onClick={() => setOccurredAt('yesterday')}
            className={`py-2 px-2 rounded-lg text-center text-xs font-bold transition-all ${dateMode === 'yesterday' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-500 hover:text-slate-900'}`}
          >
            Yesterday
          </button>
          <button 
            onClick={() => setDateMode('custom')}
            className={`py-2 px-2 rounded-lg text-center text-xs font-bold transition-all flex items-center justify-center gap-1 ${dateMode === 'custom' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-500 hover:text-slate-900'}`}
          >
            <CalendarToday size={14} />
            <span>Custom</span>
          </button>
        </div>

        {dateMode === 'custom' && (
          <div className="mt-3">
            <input 
              type="date" 
              className="w-full bg-slate-50 border-0 rounded-lg text-xs font-bold p-2 focus:ring-0 text-slate-900" 
              value={`${occurredAtDate.getFullYear()}-${String(occurredAtDate.getMonth() + 1).padStart(2, '0')}-${String(occurredAtDate.getDate()).padStart(2, '0')}`}
              onChange={(e) => setOccurredAt('custom', e.target.value)}
            />
          </div>
        )}

        {/* Time / Clock Feature */}
        <div className="mt-4 pt-3 border-t border-slate-100 flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-full bg-slate-100 flex items-center justify-center text-slate-600">
              <Schedule size={18} />
            </div>
            <div className="flex flex-col">
              <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Time</span>
              <span className="text-xs font-bold text-slate-900">
                {formatTime(safeOccurredAt)}
              </span>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button 
              type="button"
              onClick={handleSetCurrentTime}
              className="px-2.5 py-1.5 text-[11px] font-bold text-blue-600 bg-blue-50 hover:bg-blue-100 rounded-lg transition-colors active:scale-95"
            >
              Now
            </button>
            <input 
              type="time" 
              className="bg-slate-50 border border-slate-200 rounded-lg text-xs font-bold px-2 py-1.5 text-slate-900 focus:outline-none focus:ring-1 focus:ring-blue-500 cursor-pointer"
              value={formatTimeInput(occurredAtDate)}
              onChange={(e) => handleTimeChange(e.target.value)}
            />
          </div>
        </div>
      </section>

      <button
        type="button"
        role="switch"
        aria-checked={addAnother}
        onClick={() => setAddAnother((value) => !value)}
        className={`mb-3 flex w-full items-center justify-between gap-3 rounded-xl border px-3 py-2 text-left transition-colors ${
          addAnother ? 'border-blue-200 bg-blue-50' : 'border-slate-200 bg-white hover:bg-slate-50'
        }`}
      >
        <span>
          <span className="block text-xs font-bold text-slate-800">Add another transaction</span>
          <span className="block text-[10px] text-slate-500">Keep this date and time</span>
        </span>
        <span className={`relative h-5 w-9 shrink-0 rounded-full transition-colors ${addAnother ? 'bg-blue-600' : 'bg-slate-300'}`}>
          <span className={`absolute top-0.5 h-4 w-4 rounded-full bg-white shadow-sm transition-transform ${addAnother ? 'translate-x-4' : 'translate-x-0.5'}`} />
        </span>
      </button>

      {/* Actions */}
      <div className="w-full flex flex-col gap-2 mt-auto">
        <button 
          onClick={handleSave}
          disabled={loading || !formData.name.trim()}
          className={`w-full h-14 rounded-full ${typeTheme.btnBg} text-white font-bold flex items-center justify-center gap-2 shadow-lg active:scale-[0.98] transition-all disabled:cursor-not-allowed disabled:opacity-40`}
        >
          {loading ? (
            <span className="w-5 h-5 rounded-full border-2 border-white border-t-transparent animate-spin"></span>
          ) : (
            <>
              <CheckCircle size={20} />
              <span>Save Transaction</span>
            </>
          )}
        </button>
        <button onClick={onBack} className="w-full py-3 text-center text-slate-400 hover:text-slate-900 text-xs font-bold transition-colors">
          ← Back to Details
        </button>
      </div>
    </div>
  );
}
