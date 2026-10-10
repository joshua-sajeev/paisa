"use client";

import { useState } from "react";
import { 
  ArrowBack, 
  Close, 
  CheckCircle, 
  AccountBalance, 
  Restaurant, 
  Savings, 
  CalendarToday
} from "@material-symbols-svg/react/w400";
import { TransactionFormData } from "@/app/(protected)/transactions/add/page";
import { apiFetch } from "@/lib/api";
import { formatMoney, CATEGORIES } from "@/components/transactions/transaction-utils";

type Step3Props = {
  formData: TransactionFormData;
  updateFormData: (data: Partial<TransactionFormData>) => void;
  onBack: () => void;
  onCancel?: () => void;
  onComplete: () => void;
};

export default function Step3Confirmation({ formData, updateFormData, onBack, onCancel, onComplete }: Step3Props) {
  const [loading, setLoading] = useState(false);
  const getInitialDateMode = () => {
    const d = new Date(formData.occurredAt);
    const today = new Date();
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);

    if (d.toDateString() === today.toDateString()) return 'today';
    if (d.toDateString() === yesterday.toDateString()) return 'yesterday';
    return 'custom';
  };

  const [dateMode, setDateMode] = useState<'today' | 'yesterday' | 'custom'>(getInitialDateMode());

  const handleSave = async () => {
    setLoading(true);
    try {
      const payload = {
        name: formData.name || (formData.type.charAt(0).toUpperCase() + formData.type.slice(1)),
        type: formData.type,
        category: formData.category,
        from_account_id: formData.type === 'income' ? "" : formData.fromAccountId,
        to_account_id: formData.type === 'expense' ? "" : formData.toAccountId,
        jar_id: formData.jarId || "",
        amount: Math.round(parseFloat(formData.amount) * 100),
        occurred_at: formData.occurredAt,
        is_master_income: formData.isMasterIncome
      };

      await apiFetch("/api/v1/transactions/", {
        method: "POST",
        body: JSON.stringify(payload)
      });

      onComplete();
    } catch (err) {
      console.error("Failed to save transaction", err);
      alert("Failed to save transaction. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  const setOccurredAt = (mode: 'today' | 'yesterday' | 'custom', customDate?: string) => {
    setDateMode(mode);
    const date = new Date(formData.occurredAt);
    const now = new Date();
    if (mode === 'today') {
      date.setFullYear(now.getFullYear(), now.getMonth(), now.getDate());
    } else if (mode === 'yesterday') {
      now.setDate(now.getDate() - 1);
      date.setFullYear(now.getFullYear(), now.getMonth(), now.getDate());
    } else if (mode === 'custom' && customDate) {
      const [y, m, d] = customDate.split('-').map(Number);
      date.setFullYear(y, m - 1, d);
    }
    updateFormData({ occurredAt: date.toISOString() });
  };

  const categoryLabel = CATEGORIES.find(c => c.value === formData.category)?.label || 'Other';
  const currentAmount = parseFloat(formData.amount) * 100;

  const typeTheme = {
    expense: {
      stepText: 'text-red-500',
      pillBg: 'bg-red-500',
      btnBg: 'bg-red-500 hover:bg-red-600 shadow-red-500/25',
    },
    income: {
      stepText: 'text-emerald-500',
      pillBg: 'bg-emerald-500',
      btnBg: 'bg-emerald-500 hover:bg-emerald-600 shadow-emerald-500/25',
    },
    transfer: {
      stepText: 'text-blue-600',
      pillBg: 'bg-blue-600',
      btnBg: 'bg-blue-600 hover:bg-blue-700 shadow-blue-500/25',
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
      <section className="w-full bg-white rounded-xl p-8 shadow-sm flex flex-col items-center text-center relative overflow-hidden mb-6 border border-slate-100">
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
      </section>

      {/* Details */}
      <section className="w-full bg-white rounded-xl p-4 shadow-sm mb-6 border border-slate-100">
        <div className="flex items-center justify-between pb-2 mb-2 border-b border-slate-50">
          <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Ledger Details</span>
        </div>
        
        {/* Account(s) */}
        <div className="flex items-center justify-between py-3 border-b border-slate-50">
          <div className="flex items-center gap-3 min-w-0">
            <div className="w-10 h-10 rounded-full bg-slate-100 flex items-center justify-center shrink-0 text-blue-600">
              <AccountBalance size={20} />
            </div>
            <div className="flex flex-col min-w-0">
              <span className="text-[10px] font-bold text-slate-400 uppercase">{formData.type === 'income' ? 'Deposit To' : 'Payment Source'}</span>
              <span className="text-sm font-bold text-slate-900 truncate">
                {formData.type === 'income' ? formData.toAccountName : formData.fromAccountName}
              </span>
            </div>
          </div>
        </div>

        {/* Category */}
        <div className="flex items-center justify-between py-3 border-b border-slate-50">
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

        {/* Jar */}
        {formData.jarId && (
          <div className="flex items-center justify-between py-3">
            <div className="flex items-center gap-3 min-w-0">
              <div className="w-10 h-10 rounded-full bg-slate-100 flex items-center justify-center shrink-0 text-purple-600">
                <Savings size={20} />
              </div>
              <div className="flex flex-col min-w-0">
                <span className="text-[10px] font-bold text-slate-400 uppercase">Vault / Jar</span>
                <span className="text-sm font-bold text-slate-900 truncate">{formData.jarName}</span>
              </div>
            </div>
          </div>
        )}
      </section>

      {/* Date Allocation */}
      <section className="w-full bg-white rounded-xl p-4 shadow-sm mb-8 border border-slate-100">
        <div className="flex items-center justify-between mb-2">
          <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Posting Date</span>
          <span className="text-xs font-bold text-blue-600">
            {dateMode === 'today' ? 'Today' : dateMode === 'yesterday' ? 'Yesterday' : new Date(formData.occurredAt).toLocaleDateString()}
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
              className="w-full bg-slate-50 border-0 rounded-lg text-xs font-bold p-2 focus:ring-0" 
              value={`${new Date(formData.occurredAt).getFullYear()}-${String(new Date(formData.occurredAt).getMonth() + 1).padStart(2, '0')}-${String(new Date(formData.occurredAt).getDate()).padStart(2, '0')}`}
              onChange={(e) => setOccurredAt('custom', e.target.value)}
            />
          </div>
        )}
      </section>

      {/* Actions */}
      <div className="w-full flex flex-col gap-2 mt-auto">
        <button 
          onClick={handleSave}
          disabled={loading}
          className={`w-full h-14 rounded-full ${typeTheme.btnBg} text-white font-bold flex items-center justify-center gap-2 shadow-lg active:scale-[0.98] transition-all disabled:opacity-50`}
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
