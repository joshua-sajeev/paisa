"use client";

import { useState } from "react";
import { 
  Close, 
  ArrowForward, 
  Calculate, 
  TrendingDown, 
  TrendingUp, 
  SwapHoriz,
  Backspace
} from "@material-symbols-svg/react/w400";
import { TransactionFormData } from "@/app/(protected)/transactions/add/page";
import { TransactionType } from "@/components/transactions/transaction-utils";
import CalculatorModal from "@/components/transactions/CalculatorModal";
import { toAmountString } from "@/lib/calculator";

type Step1Props = {
  formData: TransactionFormData;
  updateFormData: (data: Partial<TransactionFormData>) => void;
  onNext: () => void;
  onCancel: () => void;
};

const config = {
  expense: {
    color: '#ef4444',
    accentClass: 'text-red-500',
    bgAccentClass: 'bg-red-500',
    hoverBgClass: 'hover:bg-red-600',
    shadowClass: 'shadow-red-500/25',
    sign: '-',
    activePill: 'bg-white text-red-600 border border-red-200/80 ring-2 ring-red-500/10 shadow-sm',
    icon: <TrendingDown size={20} />,
  },
  income: {
    color: '#10b981',
    accentClass: 'text-emerald-500',
    bgAccentClass: 'bg-emerald-500',
    hoverBgClass: 'hover:bg-emerald-600',
    shadowClass: 'shadow-emerald-500/25',
    sign: '+',
    activePill: 'bg-white text-emerald-600 border border-emerald-200/80 ring-2 ring-emerald-500/10 shadow-sm',
    icon: <TrendingUp size={20} />,
  },
  transfer: {
    color: '#2563eb',
    accentClass: 'text-blue-600',
    bgAccentClass: 'bg-blue-600',
    hoverBgClass: 'hover:bg-blue-700',
    shadowClass: 'shadow-blue-500/25',
    sign: '⇄',
    activePill: 'bg-white text-blue-600 border border-blue-200/80 ring-2 ring-blue-500/10 shadow-sm',
    icon: <SwapHoriz size={20} />,
  }
};

export default function Step1Amount({ formData, updateFormData, onNext, onCancel }: Step1Props) {
  const [isCalculatorOpen, setIsCalculatorOpen] = useState(false);

  const formatAmount = (val: string) => {
    if (!val || val === "0") return "0";
    const parts = val.split('.');
    const integerPart = parts[0];
    const decimalPart = parts.length > 1 ? '.' + parts[1] : '';

    let lastThree = integerPart.substring(integerPart.length - 3);
    const otherNumbers = integerPart.substring(0, integerPart.length - 3);
    if (otherNumbers !== '') {
      lastThree = ',' + lastThree;
    }
    const formattedInteger = otherNumbers.replace(/\B(?=(\d{2})+(?!\d))/g, ",") + lastThree;
    return formattedInteger + decimalPart;
  };

  const appendDigit = (d: string) => {
    let newAmount = formData.amount;
    if (newAmount === "0" && d !== "0") {
      newAmount = d;
    } else if (newAmount.length < 9) {
      if (newAmount.includes('.')) {
        const parts = newAmount.split('.');
        if (parts[1] && parts[1].length >= 2) return;
      }
      newAmount += d;
    }
    updateFormData({ amount: newAmount });
  };

  const appendDecimal = () => {
    if (!formData.amount.includes('.')) {
      const newAmount = formData.amount === "" ? "0." : formData.amount + ".";
      updateFormData({ amount: newAmount });
    }
  };

  const deleteDigit = () => {
    let newAmount = formData.amount;
    if (newAmount.length > 0) {
      newAmount = newAmount.slice(0, -1);
    }
    if (newAmount === "" || newAmount === "-") {
      newAmount = "0";
    }
    updateFormData({ amount: newAmount });
  };

  const addQuickAmount = (val: number) => {
    const currentVal = parseFloat(formData.amount) || 0;
    updateFormData({ amount: toAmountString(currentVal + val) });
  };

  const currentConfig = config[formData.type];

  return (
    <div className="flex flex-col flex-1 w-full bg-white relative overflow-hidden h-full min-h-screen">
      {/* Header */}
      <header className="px-5 py-3 flex items-center justify-between relative">
        <button onClick={onCancel} className="w-10 h-10 rounded-full flex items-center justify-center text-slate-600 hover:bg-slate-100 transition-colors">
          <Close size={24} />
        </button>
        <div className="flex items-center gap-2 bg-slate-100/80 px-3.5 py-1.5 rounded-full border border-slate-200/60">
          <span className="w-2 h-2 rounded-full bg-slate-900"></span>
          <span className="text-xs font-bold tracking-wide uppercase text-slate-700">Step 1 of 3</span>
        </div>
        <button 
          className="text-xs font-bold text-slate-400 hover:text-slate-700 transition-colors px-2 py-1"
          onClick={() => updateFormData({ amount: "0" })}
        >
          Clear
        </button>
      </header>

      {/* Main Content */}
      <main className="flex-1 flex flex-col justify-center px-6 py-4 w-full">
        <div className="flex flex-col items-center mb-8">
          <p className="text-xs font-medium uppercase tracking-widest text-slate-400 mb-2">Enter Amount</p>
          
          <div className="w-full py-2 flex flex-col items-center justify-center cursor-pointer select-none">
            <div className="flex items-center justify-center gap-1.5 sm:gap-2">
              <span className="text-3xl sm:text-4xl font-extrabold transition-colors" style={{ color: currentConfig.color }}>₹</span>
              <div className="flex items-baseline font-bold tracking-tight">
                <span className="text-5xl sm:text-6xl text-slate-900 font-extrabold transition-colors">
                  {formatAmount(formData.amount)}
                </span>
                <span className="inline-block w-1 h-10 sm:h-12 rounded-full ml-1 animate-pulse transition-colors" style={{ backgroundColor: currentConfig.color }}></span>
              </div>
            </div>
            
            <div className="mt-2 text-xs font-medium text-slate-400 flex items-center gap-1.5">
              <span className="font-bold" style={{ color: currentConfig.color }}>{currentConfig.sign}</span>
              <span>{formData.type === 'expense' ? 'Money going out' : formData.type === 'income' ? 'Money coming in' : 'Move between accounts'}</span>
            </div>

            <button 
              className="mt-2.5 inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold text-slate-600 bg-slate-100 hover:bg-slate-200/80 border border-slate-200/80 active:scale-95 transition-all shadow-sm"
              onClick={() => setIsCalculatorOpen(true)}
            >
              <Calculate size={16} className="text-slate-500" />
              <span>Calculator</span>
            </button>
          </div>
        </div>

        <div className="w-full flex flex-col gap-5 max-w-[420px] mx-auto">
          {/* Type Selector */}
          <div className="w-full p-1 bg-slate-100 rounded-2xl grid grid-cols-3 gap-1 border border-slate-200/70 select-none">
            {(['expense', 'income', 'transfer'] as TransactionType[]).map((t) => (
              <button
                key={t}
                onClick={() => updateFormData({ type: t })}
                className={`py-2.5 px-2 rounded-xl flex items-center justify-center gap-1.5 text-xs font-bold transition-all ${
                  formData.type === t ? currentConfig.activePill : 'text-slate-600 hover:text-slate-900 hover:bg-white/50'
                }`}
              >
                <span className={`w-5 h-5 rounded-full flex items-center justify-center ${formData.type === t ? '' : 'bg-slate-200'}`}>
                  {config[t].icon}
                </span>
                <span className="capitalize">{t}</span>
              </button>
            ))}
          </div>

          {/* Quick Amounts */}
          <div className="flex items-center gap-2 overflow-x-auto w-full pb-1 no-scrollbar justify-center text-xs font-semibold">
            {[10, 20, 100, 500].map(amt => (
              <button 
                key={amt}
                onClick={() => addQuickAmount(amt)}
                className="px-3 py-1.5 rounded-full bg-white border border-slate-200 text-slate-600 hover:border-slate-300 active:scale-95 transition-all"
              >
                +₹{amt}
              </button>
            ))}
          </div>

          {/* Keypad */}
          <div className="w-full">
            <div className="grid grid-cols-3 gap-y-2 gap-x-3 text-center select-none max-w-[340px] mx-auto">
              {['1', '2', '3', '4', '5', '6', '7', '8', '9'].map(d => (
                <button key={d} onClick={() => appendDigit(d)} className="h-14 rounded-2xl flex items-center justify-center text-2xl font-semibold text-slate-800 hover:bg-slate-100 transition-all active:scale-95">
                  {d}
                </button>
              ))}
              <button onClick={appendDecimal} className="h-14 rounded-2xl flex items-center justify-center text-2xl font-bold text-slate-800 hover:bg-slate-100 transition-all active:scale-95">•</button>
              <button onClick={() => appendDigit('0')} className="h-14 rounded-2xl flex items-center justify-center text-2xl font-semibold text-slate-800 hover:bg-slate-100 transition-all active:scale-95">0</button>
              <button onClick={deleteDigit} className="h-14 rounded-2xl flex items-center justify-center text-xl text-slate-700 hover:bg-slate-100 active:text-red-500 transition-all active:scale-95">
                <Backspace size={24} />
              </button>
            </div>

            <div className="mt-4 pt-1">
              <button 
                onClick={onNext}
                disabled={parseFloat(formData.amount) <= 0}
                className={`w-full py-4 px-6 rounded-2xl font-bold text-white shadow-lg flex items-center justify-center gap-2.5 transition-all duration-200 active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed ${currentConfig.bgAccentClass} ${currentConfig.hoverBgClass} ${currentConfig.shadowClass}`}
              >
                <span>Continue to Details</span>
                <ArrowForward size={18} />
              </button>
            </div>
          </div>
        </div>
      </main>

      {/* iOS Home Indicator */}
      <div className="w-full pb-2 pt-1 flex justify-center select-none">
        <div className="w-32 h-1 bg-slate-300 rounded-full"></div>
      </div>

      {/* Calculator Modal */}
      {isCalculatorOpen && (
        <CalculatorModal
          initialAmount={formData.amount}
          accent={{
            color: currentConfig.color,
            bgClass: currentConfig.bgAccentClass,
            hoverClass: currentConfig.hoverBgClass,
            shadowClass: currentConfig.shadowClass,
          }}
          onClose={() => setIsCalculatorOpen(false)}
          onApply={(amount) => {
            updateFormData({ amount });
            setIsCalculatorOpen(false);
          }}
        />
      )}
    </div>
  );
}
