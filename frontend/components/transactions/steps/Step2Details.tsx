"use client";

import { useState, useEffect } from "react";
import { 
  ArrowBack, 
  Payments, 
  ReceiptLong,
  AccountBalanceWallet, 
  Category as CategoryIcon, 
  Savings, 
  EditNote, 
  CheckCircle, 
  ArrowForward,
  Check,
  Restaurant,
  DirectionsCar,
  Movie,
  ShoppingCart,
  Favorite,
  SyncAlt,
  VolunteerActivism,
  TrendingUp,
  Home,
  MoreHoriz
} from "@material-symbols-svg/react/w400";
import { TransactionFormData } from "@/app/(protected)/transactions/add/page";
import { Account, getAccounts } from "@/lib/accounts/accounts";
import { apiFetch } from "@/lib/api";
import { CATEGORIES, formatMoney } from "@/components/transactions/transaction-utils";

type Jar = {
  id: string;
  name: string;
  balance: number;
  goal_amount: number;
  icon_key: string | null;
};

type Step2Props = {
  formData: TransactionFormData;
  updateFormData: (data: Partial<TransactionFormData>) => void;
  onNext: () => void;
  onBack: () => void;
};

const CATEGORY_ICONS: Record<string, any> = {
  food: Restaurant,
  transport: DirectionsCar,
  entertainment: Movie,
  groceries: ShoppingCart,
  health: Favorite,
  transfer: SyncAlt,
  donation: VolunteerActivism,
  investment: TrendingUp,
  housing: Home,
  other: MoreHoriz,
};

const getJarIcon = (name: string) => {
  const key = name.toLowerCase();
  if (key.includes("necessit")) return ReceiptLong;
  if (key.includes("leisure")) return Movie;
  if (key.includes("invest")) return TrendingUp;
  if (key.includes("giv")) return VolunteerActivism;
  return Savings;
};

export default function Step2Details({ formData, updateFormData, onNext, onBack }: Step2Props) {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [jars, setJars] = useState<Jar[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function loadData() {
      try {
        const [accts, allocResp] = await Promise.all([
          getAccounts(),
          apiFetch<{ allocations: any[] }>("/api/v1/allocations/").catch(() => null)
        ]);
        if (accts) setAccounts(accts);
        
        if (allocResp && allocResp.allocations) {
          const jarMap = new Map<string, Jar>();
          for (const alloc of allocResp.allocations) {
            const existing = jarMap.get(alloc.jar_id);
            if (existing) {
              existing.balance += alloc.amount;
            } else {
              jarMap.set(alloc.jar_id, {
                id: alloc.jar_id,
                name: alloc.jar_name,
                balance: alloc.amount,
                goal_amount: 10000000, // mock goal to avoid divide by zero and show some progress
                icon_key: null,
              });
            }
          }
          setJars(Array.from(jarMap.values()));
        }

        // Set default accounts if not set
        if (accts && accts.length > 0) {
          const primary = accts.find(a => a.is_primary) || accts[0];
          if (!formData.fromAccountId && (formData.type === 'expense' || formData.type === 'transfer')) {
            updateFormData({ fromAccountId: primary.id, fromAccountName: primary.name });
          }
          if (!formData.toAccountId && (formData.type === 'income' || formData.type === 'transfer')) {
            updateFormData({ toAccountId: primary.id, toAccountName: primary.name });
          }
        }
      } catch (err) {
        console.error("Failed to load Step 2 data", err);
      } finally {
        setLoading(false);
      }
    }
    loadData();
  }, []);

  const currentAmount = parseFloat(formData.amount) * 100; // to paise

  return (
    <div className="flex flex-col flex-1 w-full bg-white min-h-screen relative overflow-y-auto no-scrollbar">
      {/* Header */}
      <nav className="px-5 pt-4 pb-3 flex flex-col gap-3 bg-white z-20 border-b border-slate-100 sticky top-0">
        <div className="flex items-center justify-between">
          <button onClick={onBack} className="w-9 h-9 flex items-center justify-center rounded-xl bg-slate-50 border border-slate-200/80 hover:bg-slate-100 text-slate-800 transition-all active:scale-95 shadow-sm">
            <ArrowBack size={20} />
          </button>
          <div className="flex items-center gap-1.5 px-3.5 py-1 rounded-full bg-slate-900 text-white shadow-sm shadow-slate-900/10">
            <span className="w-1.5 h-1.5 rounded-full bg-red-400 animate-pulse"></span>
            <span className="text-[11px] font-extrabold tracking-wider uppercase">STEP 2 OF 3</span>
          </div>
          <button className="text-xs font-bold text-slate-400 hover:text-slate-700 transition-colors px-2 py-1 rounded-lg hover:bg-slate-50">
            Cancel
          </button>
        </div>
        
        <div className="flex items-center justify-between bg-gradient-to-r from-red-500/10 via-rose-50 to-orange-50/60 p-3 rounded-2xl border border-red-100/80">
          <div className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-xl bg-red-500 text-white flex items-center justify-center shadow-md shadow-red-500/20">
              <Payments size={18} />
            </div>
            <div>
              <p className="text-[10px] font-bold text-red-700 tracking-wider uppercase">Amount to log</p>
              <p className="text-lg font-black text-slate-900 tracking-tight leading-none">{formatMoney(currentAmount)}</p>
            </div>
          </div>
          <div className="text-right flex items-center gap-1 bg-white/80 backdrop-blur-sm border border-red-200/60 px-2.5 py-1 rounded-full">
            <span className="w-2 h-2 rounded-full bg-red-500"></span>
            <span className="text-[11px] font-bold text-slate-700 capitalize">{formData.type}</span>
          </div>
        </div>
      </nav>

      <div className="flex-1 px-5 pt-3 pb-32 overflow-y-auto no-scrollbar space-y-6">

        {/* Account Selection */}
        <section className="space-y-3">
          <div className="flex items-center justify-between px-0.5">
            <div className="flex items-center gap-1.5">
              <AccountBalanceWallet size={16} className="text-slate-400" />
              <h2 className="text-[11px] uppercase tracking-wider font-extrabold text-slate-400">
                {formData.type === 'expense' ? 'Pay From Account' : formData.type === 'income' ? 'Receive To Account' : 'From Account'}
              </h2>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-2.5">
            {accounts.map(account => (
              <div 
                key={account.id}
                onClick={() => updateFormData({ fromAccountId: account.id, fromAccountName: account.name })}
                className={`group relative p-3.5 rounded-2xl cursor-pointer flex flex-col justify-between transition-all duration-200 active:scale-[0.98] ${
                  formData.fromAccountId === account.id 
                    ? 'border-2 border-red-500 bg-gradient-to-br from-red-50/70 via-white to-rose-50/40 shadow-md shadow-red-500/10' 
                    : 'border border-slate-200/90 bg-white hover:border-slate-300'
                }`}
              >
                <div className="flex items-start justify-between mb-3">
                  <div className={`w-8 h-8 rounded-xl text-white font-black text-[11px] flex items-center justify-center shadow-sm tracking-tighter ${account.is_primary ? 'bg-blue-900' : 'bg-slate-500'}`}>
                    {account.name.substring(0, 4).toUpperCase()}
                  </div>
                  {formData.fromAccountId === account.id && (
                    <div className="w-5 h-5 rounded-full bg-red-500 text-white flex items-center justify-center shadow-sm shadow-red-500/30">
                      <Check size={13} className="font-bold" />
                    </div>
                  )}
                </div>
                <div>
                  <p className="font-bold text-slate-900 text-xs tracking-tight truncate">{account.name}</p>
                  <p className="mt-1 text-sm font-extrabold text-slate-900 font-mono tracking-tight">{formatMoney(account.balance)}</p>
                </div>
              </div>
            ))}
          </div>
        </section>

        {formData.type === 'transfer' && (
          <section className="space-y-3">
            <div className="flex items-center justify-between px-0.5">
              <div className="flex items-center gap-1.5">
                <AccountBalanceWallet size={16} className="text-slate-400" />
                <h2 className="text-[11px] uppercase tracking-wider font-extrabold text-slate-400">To Account</h2>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-2.5">
              {accounts.map(account => (
                <div 
                  key={account.id}
                  onClick={() => updateFormData({ toAccountId: account.id, toAccountName: account.name })}
                  className={`group relative p-3.5 rounded-2xl cursor-pointer flex flex-col justify-between transition-all duration-200 active:scale-[0.98] ${
                    formData.toAccountId === account.id 
                      ? 'border-2 border-blue-500 bg-gradient-to-br from-blue-50/70 via-white to-blue-50/40 shadow-md shadow-blue-500/10' 
                      : 'border border-slate-200/90 bg-white hover:border-slate-300'
                  }`}
                >
                  <div className="flex items-start justify-between mb-3">
                    <div className={`w-8 h-8 rounded-xl text-white font-black text-[11px] flex items-center justify-center shadow-sm tracking-tighter ${account.is_primary ? 'bg-blue-900' : 'bg-slate-500'}`}>
                      {account.name.substring(0, 4).toUpperCase()}
                    </div>
                    {formData.toAccountId === account.id && (
                      <div className="w-5 h-5 rounded-full bg-blue-500 text-white flex items-center justify-center shadow-sm shadow-blue-500/30">
                        <Check size={13} className="font-bold" />
                      </div>
                    )}
                  </div>
                  <div>
                    <p className="font-bold text-slate-900 text-xs tracking-tight truncate">{account.name}</p>
                    <p className="mt-1 text-sm font-extrabold text-slate-900 font-mono tracking-tight">{formatMoney(account.balance)}</p>
                  </div>
                </div>
              ))}
            </div>
          </section>
        )}

        {/* Category Selection */}
        <section className="space-y-3">
          <div className="flex items-center justify-between px-0.5">
            <div className="flex items-center gap-1.5">
              <CategoryIcon size={16} className="text-slate-400" />
              <h2 className="text-[11px] uppercase tracking-wider font-extrabold text-slate-400">Category</h2>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-2">
            {CATEGORIES.filter(c => c.value !== 'all').map(cat => {
              const Icon = CATEGORY_ICONS[cat.value] || MoreHoriz;
              return (
                <button 
                  key={cat.value}
                  onClick={() => updateFormData({ category: cat.value as any })}
                  className={`flex items-center gap-2.5 p-2.5 rounded-xl text-left transition-all active:scale-95 border ${
                    formData.category === cat.value 
                      ? 'border-2 border-red-500 bg-red-50/50 text-red-600 font-bold' 
                      : 'border-slate-200 bg-white text-slate-700'
                  }`}
                >
                  <div className={`w-8 h-8 rounded-lg flex items-center justify-center shrink-0 ${formData.category === cat.value ? 'bg-red-500 text-white' : 'bg-slate-100 text-slate-600'}`}>
                    <Icon size={18} />
                  </div>
                  <span className="font-bold text-xs truncate">{cat.label}</span>
                  {formData.category === cat.value && <CheckCircle size={16} className="ml-auto" />}
                </button>
              );
            })}
          </div>
        </section>

        {/* Jar Allocation */}
        <section className="space-y-3 pb-2">
          <div className="flex items-center justify-between px-0.5">
            <div className="flex items-center gap-1.5">
              <Savings size={16} className="text-slate-400" />
              <h2 className="text-[11px] uppercase tracking-wider font-extrabold text-slate-400">Assign To Jar</h2>
            </div>
            <span className="text-[11px] font-bold text-slate-500 bg-slate-100 px-2 py-0.5 rounded-full">Optional</span>
          </div>
          <div className="grid grid-cols-2 gap-2.5">
            {jars.map(jar => (
              <div 
                key={jar.id}
                onClick={() => updateFormData({ jarId: formData.jarId === jar.id ? '' : jar.id, jarName: formData.jarId === jar.id ? '' : jar.name })}
                className={`p-3 rounded-2xl border cursor-pointer flex flex-col justify-between transition-all active:scale-95 shadow-sm ${
                  formData.jarId === jar.id 
                    ? 'border-2 border-red-500 bg-gradient-to-br from-red-50/50 via-white to-rose-50/30' 
                    : 'border-slate-200 bg-white hover:border-slate-300'
                }`}
              >
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-1.5 min-w-0">
                    
                    <div className="w-6 h-6 rounded-full bg-slate-100 text-slate-500 flex items-center justify-center shrink-0">
                      {(() => {
                        const Icon = getJarIcon(jar.name);
                        return <Icon size={14} />;
                      })()}
                    </div>
                    <span className="text-xs font-bold text-slate-900 leading-tight truncate">{jar.name}</span>
                  </div>
                  {formData.jarId === jar.id && (
                    <div className="w-4 h-4 rounded-full bg-red-500 text-white flex items-center justify-center text-[10px]">
                      <Check size={12} className="font-bold" />
                    </div>
                  )}
                </div>
                <div className="space-y-1">
                  <div className="flex justify-between text-[10px] font-bold">
                    <span className="text-slate-700 font-mono">{formatMoney(jar.balance)}</span>
                  </div>
                  <div className="w-full h-1 bg-slate-100 rounded-full overflow-hidden">
                    <div 
                      className="h-full bg-red-500 rounded-full" 
                      style={{ width: `${Math.min(100, (jar.balance / (jar.goal_amount || 1)) * 100)}%` }}
                    ></div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </section>
      </div>

      {/* Bottom Button */}
      <div className="absolute bottom-0 inset-x-0 bg-white/95 backdrop-blur-md border-t border-slate-100 px-5 pt-3 pb-6 z-30 flex flex-col items-center">
        <button 
          onClick={onNext}
          className="tap-target w-full py-3.5 px-5 bg-gradient-to-r from-red-500 via-rose-500 to-red-600 hover:from-red-600 hover:to-rose-700 active:scale-[0.98] text-white font-extrabold text-sm rounded-2xl shadow-xl shadow-red-500/30 flex items-center justify-between transition-all duration-200 group border border-red-400/30"
        >
          <div className="flex items-center gap-2.5 text-left">
            <div className="w-8 h-8 rounded-xl bg-white/20 flex items-center justify-center backdrop-blur-sm">
              <CheckCircle size={18} />
            </div>
            <div>
              <span className="block text-[10px] font-bold text-red-100 uppercase tracking-wider leading-tight">Confirm Details</span>
              <span className="block text-sm font-black tracking-tight leading-tight">
                {formData.type === 'expense' ? 'Pay' : formData.type === 'income' ? 'Receive' : 'Transfer'} {formatMoney(currentAmount)}
              </span>
            </div>
          </div>
          <div className="flex items-center gap-1 bg-white/15 px-3 py-1.5 rounded-xl group-hover:translate-x-0.5 transition-transform">
            <span className="text-xs font-bold">Continue</span>
            <ArrowForward size={16} />
          </div>
        </button>
        <div className="w-32 h-1 bg-slate-300 rounded-full mt-3.5"></div>
      </div>
    </div>
  );
}
