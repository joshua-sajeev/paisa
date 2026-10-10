"use client";

import { useState, useEffect, type ComponentType, type SVGProps } from "react";
import { 
  ArrowBack, 
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
  TrendingDown,
  Home,
  MoreHoriz
} from "@material-symbols-svg/react/w400";
import { TransactionFormData } from "@/app/(protected)/transactions/add/page";
import { Account, getAccounts } from "@/lib/accounts/accounts";
import { apiFetch } from "@/lib/api";
import { CATEGORIES, formatMoney, TransactionCategory } from "@/components/transactions/transaction-utils";

type Jar = {
  id: string;
  name: string;
  balance: number;
  goal_amount: number;
  icon_key: string | null;
};

type AllocationItem = {
  jar_id: string;
  jar_name: string;
  amount: number;
};

type Step2Props = {
  formData: TransactionFormData;
  updateFormData: (data: Partial<TransactionFormData>) => void;
  onNext: () => void;
  onBack: () => void;
  onCancel: () => void;
};

type IconComponent = ComponentType<SVGProps<SVGSVGElement> & { size?: number }>;

const CATEGORY_ICONS: Record<string, IconComponent> = {
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

const getJarIcon = (name: string): IconComponent => {
  const key = name.toLowerCase();
  if (key.includes("necessit")) return ReceiptLong;
  if (key.includes("leisure")) return Movie;
  if (key.includes("invest")) return TrendingUp;
  if (key.includes("giv")) return VolunteerActivism;
  return Savings;
};

const THEME_CONFIG = {
  expense: {
    title: "Expense",
    actionVerb: "Pay",
    mainAccountTitle: "Pay From Account",
    headerIcon: TrendingDown,
    pulseDot: "bg-red-400",
    bannerGradient: "from-red-500/10 via-rose-50 to-orange-50/60",
    bannerBorder: "border-red-100/80",
    bannerIconBg: "bg-red-500 text-white shadow-md shadow-red-500/20",
    bannerAmountLabel: "text-red-700",
    bannerTagBorder: "border-red-200/60",
    bannerTagDot: "bg-red-500",
    bannerTagText: "text-red-700",
    noteFocusRing: "focus-within:border-red-500 focus-within:ring-2 focus-within:ring-red-500/10",
    countBadge: "text-red-600 bg-red-50 border-red-100",
    accountSelectedCard: "border-2 border-red-500 bg-gradient-to-br from-red-50/70 via-white to-rose-50/40 shadow-md shadow-red-500/10",
    accountCheckBg: "bg-red-500 text-white shadow-sm shadow-red-500/30",
    accountPrimaryTag: "text-red-600 bg-red-100/60",
    accountBottomGlow: "via-red-500",
    categorySelected: "border-2 border-red-500 bg-gradient-to-r from-red-50 via-rose-50/50 to-white text-red-600 font-bold shadow-sm shadow-red-500/10",
    categoryIconSelected: "bg-red-500 text-white shadow-sm shadow-red-500/20",
    categoryCheck: "text-red-500",
    jarSelected: "border-2 border-red-500 bg-gradient-to-br from-red-50/50 via-white to-rose-50/30 shadow-sm shadow-red-500/10",
    jarCheck: "bg-red-500 text-white shadow-sm shadow-red-500/20",
    jarProgress: "bg-red-500",
    jarBalanceText: "text-red-600",
    ctaGradient: "bg-gradient-to-r from-red-500 via-rose-500 to-red-600 hover:from-red-600 hover:to-rose-700 shadow-xl shadow-red-500/30 border border-red-400/30",
    ctaSubtext: "text-red-100",
  },
  income: {
    title: "Income",
    actionVerb: "Receive",
    mainAccountTitle: "Receive To Account",
    headerIcon: TrendingUp,
    pulseDot: "bg-emerald-400",
    bannerGradient: "from-emerald-500/10 via-emerald-50 to-teal-50/60",
    bannerBorder: "border-emerald-100/80",
    bannerIconBg: "bg-emerald-500 text-white shadow-md shadow-emerald-500/20",
    bannerAmountLabel: "text-emerald-700",
    bannerTagBorder: "border-emerald-200/60",
    bannerTagDot: "bg-emerald-500",
    bannerTagText: "text-emerald-700",
    noteFocusRing: "focus-within:border-emerald-500 focus-within:ring-2 focus-within:ring-emerald-500/10",
    countBadge: "text-emerald-600 bg-emerald-50 border-emerald-100",
    accountSelectedCard: "border-2 border-emerald-500 bg-gradient-to-br from-emerald-50/70 via-white to-teal-50/40 shadow-md shadow-emerald-500/10",
    accountCheckBg: "bg-emerald-500 text-white shadow-sm shadow-emerald-500/30",
    accountPrimaryTag: "text-emerald-600 bg-emerald-100/60",
    accountBottomGlow: "via-emerald-500",
    categorySelected: "border-2 border-emerald-500 bg-gradient-to-r from-emerald-50 via-teal-50/50 to-white text-emerald-600 font-bold shadow-sm shadow-emerald-500/10",
    categoryIconSelected: "bg-emerald-500 text-white shadow-sm shadow-emerald-500/20",
    categoryCheck: "text-emerald-500",
    jarSelected: "border-2 border-emerald-500 bg-gradient-to-br from-emerald-50/50 via-white to-teal-50/30 shadow-sm shadow-emerald-500/10",
    jarCheck: "bg-emerald-500 text-white shadow-sm shadow-emerald-500/20",
    jarProgress: "bg-emerald-500",
    jarBalanceText: "text-emerald-600",
    ctaGradient: "bg-gradient-to-r from-emerald-500 via-teal-500 to-emerald-600 hover:from-emerald-600 hover:to-teal-700 shadow-xl shadow-emerald-500/30 border border-emerald-400/30",
    ctaSubtext: "text-emerald-100",
  },
  transfer: {
    title: "Transfer",
    actionVerb: "Transfer",
    mainAccountTitle: "From Account",
    headerIcon: SyncAlt,
    pulseDot: "bg-blue-400",
    bannerGradient: "from-blue-500/10 via-blue-50 to-indigo-50/60",
    bannerBorder: "border-blue-100/80",
    bannerIconBg: "bg-blue-600 text-white shadow-md shadow-blue-500/20",
    bannerAmountLabel: "text-blue-700",
    bannerTagBorder: "border-blue-200/60",
    bannerTagDot: "bg-blue-600",
    bannerTagText: "text-blue-700",
    noteFocusRing: "focus-within:border-blue-500 focus-within:ring-2 focus-within:ring-blue-500/10",
    countBadge: "text-blue-600 bg-blue-50 border-blue-100",
    accountSelectedCard: "border-2 border-blue-500 bg-gradient-to-br from-blue-50/70 via-white to-indigo-50/40 shadow-md shadow-blue-500/10",
    accountCheckBg: "bg-blue-600 text-white shadow-sm shadow-blue-500/30",
    accountPrimaryTag: "text-blue-600 bg-blue-100/60",
    accountBottomGlow: "via-blue-500",
    categorySelected: "border-2 border-blue-500 bg-gradient-to-r from-blue-50 via-indigo-50/50 to-white text-blue-600 font-bold shadow-sm shadow-blue-500/10",
    categoryIconSelected: "bg-blue-600 text-white shadow-sm shadow-blue-500/20",
    categoryCheck: "text-blue-600",
    jarSelected: "border-2 border-blue-500 bg-gradient-to-br from-blue-50/50 via-white to-indigo-50/30 shadow-sm shadow-blue-500/10",
    jarCheck: "bg-blue-600 text-white shadow-sm shadow-blue-500/20",
    jarProgress: "bg-blue-600",
    jarBalanceText: "text-blue-600",
    ctaGradient: "bg-gradient-to-r from-blue-600 via-indigo-600 to-blue-700 hover:from-blue-700 hover:to-indigo-800 shadow-xl shadow-blue-500/30 border border-blue-400/30",
    ctaSubtext: "text-blue-100",
  }
};

export default function Step2Details({ formData, updateFormData, onNext, onBack, onCancel }: Step2Props) {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [jars, setJars] = useState<Jar[]>([]);
  const [loading, setLoading] = useState(true);

  const theme = THEME_CONFIG[formData.type];
  const HeaderTypeIcon = theme.headerIcon;

  useEffect(() => {
    let mounted = true;
    async function loadData() {
      try {
        const [accts, allocResp] = await Promise.all([
          getAccounts(),
          apiFetch<{ allocations: AllocationItem[] }>("/api/v1/allocations/").catch(() => null)
        ]);
        if (!mounted) return;
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
                goal_amount: 10000000,
                icon_key: null,
              });
            }
          }
          setJars(Array.from(jarMap.values()));
        }

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
        if (mounted) setLoading(false);
      }
    }
    loadData();
    return () => {
      mounted = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const currentAmount = parseFloat(formData.amount) * 100; // to paise

  const isPrimarySelected = (account: Account) => {
    if (formData.type === 'income') {
      return formData.toAccountId === account.id;
    }
    return formData.fromAccountId === account.id;
  };

  const handlePrimaryAccountSelect = (account: Account) => {
    if (formData.type === 'income') {
      updateFormData({ toAccountId: account.id, toAccountName: account.name });
    } else {
      updateFormData({ fromAccountId: account.id, fromAccountName: account.name });
    }
  };

  return (
    <div className="flex flex-col flex-1 w-full bg-white min-h-screen relative overflow-y-auto no-scrollbar">
      {/* Header */}
      <nav className="px-5 pt-4 pb-3 flex flex-col gap-3 bg-white z-20 border-b border-slate-100 sticky top-0">
        <div className="flex items-center justify-between">
          <button 
            onClick={onBack} 
            className="w-9 h-9 flex items-center justify-center rounded-xl bg-slate-50 border border-slate-200/80 hover:bg-slate-100 text-slate-800 transition-all active:scale-95 shadow-sm"
            aria-label="Back to amount"
          >
            <ArrowBack size={20} />
          </button>
          <div className="flex items-center gap-1.5 px-3.5 py-1 rounded-full bg-slate-900 text-white shadow-sm shadow-slate-900/10">
            <span className={`w-1.5 h-1.5 rounded-full ${theme.pulseDot} animate-pulse`}></span>
            <span className="text-[11px] font-extrabold tracking-wider uppercase">STEP 2 OF 3</span>
          </div>
          <button 
            onClick={onCancel}
            className="text-xs font-bold text-slate-400 hover:text-slate-700 transition-colors px-2 py-1 rounded-lg hover:bg-slate-50"
          >
            Cancel
          </button>
        </div>
        
        {/* Continuity Banner */}
        <div className={`flex items-center justify-between bg-gradient-to-r ${theme.bannerGradient} p-3 rounded-2xl border ${theme.bannerBorder} transition-colors duration-300`}>
          <div className="flex items-center gap-2.5">
            <div className={`w-8 h-8 rounded-xl ${theme.bannerIconBg} flex items-center justify-center transition-all duration-300`}>
              <HeaderTypeIcon size={18} />
            </div>
            <div>
              <p className={`text-[10px] font-bold ${theme.bannerAmountLabel} tracking-wider uppercase`}>Amount to log</p>
              <p className="text-lg font-black text-slate-900 tracking-tight leading-none">{formatMoney(currentAmount)}</p>
            </div>
          </div>
          <div className={`text-right flex items-center gap-1.5 bg-white/85 backdrop-blur-sm border ${theme.bannerTagBorder} px-2.5 py-1 rounded-full shadow-xs`}>
            <span className={`w-2 h-2 rounded-full ${theme.bannerTagDot}`}></span>
            <span className={`text-[11px] font-bold capitalize ${theme.bannerTagText}`}>{theme.title}</span>
          </div>
        </div>
      </nav>

      <div className="flex-1 px-5 pt-3 pb-32 overflow-y-auto no-scrollbar space-y-6">

        {/* Note / Title Input */}
        {/* <div className={`p-3 rounded-2xl bg-white border border-slate-200/90 shadow-xs flex items-center justify-between gap-3 ${theme.noteFocusRing} transition-all`}>
          <div className="flex items-center gap-2.5 flex-1 min-w-0">
            <div className="w-8 h-8 rounded-xl bg-slate-50 border border-slate-200/80 text-slate-500 flex items-center justify-center shrink-0">
              <EditNote size={18} />
            </div>
            <div className="flex-1 min-w-0">
              <label htmlFor="tx-note-input" className="block text-[10px] font-extrabold uppercase tracking-wider text-slate-400 leading-none mb-1">
                Note / Title
              </label>
              <input 
                id="tx-note-input" 
                type="text" 
                value={formData.name}
                onChange={(e) => updateFormData({ name: e.target.value })}
                placeholder={
                  formData.type === 'expense' 
                    ? "Add transaction note (e.g. Dinner, Groceries)..." 
                    : formData.type === 'income' 
                    ? "Add income source (e.g. Salary, Dividend)..." 
                    : "Add transfer description..."
                }
                className="w-full p-0 border-0 text-xs font-bold text-slate-900 bg-transparent focus:ring-0 focus:outline-none placeholder:text-slate-400 placeholder:font-normal leading-tight" 
              />
            </div>
          </div>
        </div> */}

        {/* Account Selection */}
        <section className="space-y-3">
          <div className="flex items-center justify-between px-0.5">
            <div className="flex items-center gap-1.5">
              <AccountBalanceWallet size={16} className="text-slate-400" />
              <h2 className="text-[11px] uppercase tracking-wider font-extrabold text-slate-400">
                {theme.mainAccountTitle}
              </h2>
            </div>
            {accounts.length > 0 && (
              <span className={`text-[11px] font-bold px-2 py-0.5 rounded-full border ${theme.countBadge}`}>
                {accounts.length} Available
              </span>
            )}
          </div>

          {loading ? (
            <div className="grid grid-cols-2 gap-2.5">
              {[1, 2].map((i) => (
                <div key={i} className="h-24 rounded-2xl bg-slate-100 animate-pulse border border-slate-200" />
              ))}
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-2.5">
              {accounts.map(account => {
                const selected = isPrimarySelected(account);
                return (
                  <div 
                    key={account.id}
                    onClick={() => handlePrimaryAccountSelect(account)}
                    className={`group relative p-3.5 rounded-2xl cursor-pointer flex flex-col justify-between transition-all duration-200 active:scale-[0.98] ${
                      selected 
                        ? theme.accountSelectedCard 
                        : 'border border-slate-200/90 bg-white hover:border-slate-300 hover:shadow-xs'
                    }`}
                  >
                    <div className="flex items-start justify-between mb-3">
                      <div className={`w-8 h-8 rounded-xl text-white font-black text-[11px] flex items-center justify-center shadow-xs tracking-tighter ${account.is_primary ? 'bg-blue-900' : 'bg-slate-500'}`}>
                        {account.name.substring(0, 4).toUpperCase()}
                      </div>
                      {selected ? (
                        <div className={`w-5 h-5 rounded-full flex items-center justify-center ${theme.accountCheckBg}`}>
                          <Check size={13} className="font-bold" />
                        </div>
                      ) : (
                        <div className="w-4 h-4 rounded-full border border-slate-300 group-hover:border-slate-400" />
                      )}
                    </div>
                    <div>
                      <div className="flex items-center justify-between gap-1">
                        <p className="font-bold text-slate-900 text-xs tracking-tight truncate">{account.name}</p>
                        {account.is_primary && (
                          <span className={`text-[9px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded-md shrink-0 ${theme.accountPrimaryTag}`}>
                            Primary
                          </span>
                        )}
                      </div>
                      <p className="mt-1 text-sm font-extrabold text-slate-900 font-mono tracking-tight">{formatMoney(account.balance)}</p>
                    </div>
                    {selected && (
                      <div className={`absolute -bottom-px inset-x-4 h-0.5 bg-gradient-to-r from-transparent ${theme.accountBottomGlow} to-transparent`}></div>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </section>

        {/* Transfer: Destination Account Selection */}
        {formData.type === 'transfer' && (
          <section className="space-y-3">
            <div className="flex items-center justify-between px-0.5">
              <div className="flex items-center gap-1.5">
                <AccountBalanceWallet size={16} className="text-slate-400" />
                <h2 className="text-[11px] uppercase tracking-wider font-extrabold text-slate-400">To Account</h2>
              </div>
              <span className="text-[11px] font-bold text-indigo-600 bg-indigo-50 border border-indigo-100 px-2 py-0.5 rounded-full">
                Destination
              </span>
            </div>
            <div className="grid grid-cols-2 gap-2.5">
              {accounts.map(account => {
                const isSelectedTo = formData.toAccountId === account.id;
                return (
                  <div 
                    key={account.id}
                    onClick={() => updateFormData({ toAccountId: account.id, toAccountName: account.name })}
                    className={`group relative p-3.5 rounded-2xl cursor-pointer flex flex-col justify-between transition-all duration-200 active:scale-[0.98] ${
                      isSelectedTo 
                        ? 'border-2 border-indigo-500 bg-gradient-to-br from-indigo-50/70 via-white to-blue-50/40 shadow-md shadow-indigo-500/10' 
                        : 'border border-slate-200/90 bg-white hover:border-slate-300 hover:shadow-xs'
                    }`}
                  >
                    <div className="flex items-start justify-between mb-3">
                      <div className={`w-8 h-8 rounded-xl text-white font-black text-[11px] flex items-center justify-center shadow-xs tracking-tighter ${account.is_primary ? 'bg-blue-900' : 'bg-slate-500'}`}>
                        {account.name.substring(0, 4).toUpperCase()}
                      </div>
                      {isSelectedTo ? (
                        <div className="w-5 h-5 rounded-full bg-indigo-600 text-white flex items-center justify-center shadow-sm shadow-indigo-500/30">
                          <Check size={13} className="font-bold" />
                        </div>
                      ) : (
                        <div className="w-4 h-4 rounded-full border border-slate-300 group-hover:border-slate-400" />
                      )}
                    </div>
                    <div>
                      <div className="flex items-center justify-between gap-1">
                        <p className="font-bold text-slate-900 text-xs tracking-tight truncate">{account.name}</p>
                        {account.is_primary && (
                          <span className="text-[9px] font-bold uppercase tracking-wider text-indigo-600 bg-indigo-100/60 px-1.5 py-0.5 rounded-md shrink-0">
                            Primary
                          </span>
                        )}
                      </div>
                      <p className="mt-1 text-sm font-extrabold text-slate-900 font-mono tracking-tight">{formatMoney(account.balance)}</p>
                    </div>
                    {isSelectedTo && (
                      <div className="absolute -bottom-px inset-x-4 h-0.5 bg-gradient-to-r from-transparent via-indigo-500 to-transparent"></div>
                    )}
                  </div>
                );
              })}
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
            <span className="text-[11px] font-bold text-slate-500 bg-slate-100 px-2 py-0.5 rounded-full">
              {CATEGORIES.find(c => c.value === formData.category)?.label || "Selected"}
            </span>
          </div>
          <div className="grid grid-cols-2 gap-2">
            {CATEGORIES.filter(c => c.value !== 'all').map(cat => {
              const Icon = CATEGORY_ICONS[cat.value] || MoreHoriz;
              const isSelected = formData.category === cat.value;
              return (
                <button 
                  key={cat.value}
                  type="button"
                  onClick={() => updateFormData({ category: cat.value as TransactionCategory })}
                  className={`cat-pill flex items-center gap-2.5 p-2.5 rounded-xl text-left transition-all active:scale-95 border ${
                    isSelected 
                      ? theme.categorySelected 
                      : 'border-slate-200 bg-white hover:border-slate-300 hover:bg-slate-50 text-slate-700'
                  }`}
                >
                  <div className={`w-8 h-8 rounded-lg flex items-center justify-center shrink-0 transition-transform group-hover:scale-105 ${
                    isSelected ? theme.categoryIconSelected : 'bg-slate-100 text-slate-600'
                  }`}>
                    <Icon size={18} />
                  </div>
                  <span className="font-bold text-xs truncate flex-1">{cat.label}</span>
                  {isSelected && <CheckCircle size={16} className={`ml-auto shrink-0 ${theme.categoryCheck}`} />}
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
            {jars.map(jar => {
              const isSelectedJar = formData.jarId === jar.id;
              const progressPct = Math.min(100, Math.max(0, (jar.balance / (jar.goal_amount || 1)) * 100));
              return (
                <div 
                  key={jar.id}
                  onClick={() => updateFormData({ 
                    jarId: isSelectedJar ? '' : jar.id, 
                    jarName: isSelectedJar ? '' : jar.name 
                  })}
                  className={`p-3 rounded-2xl border cursor-pointer flex flex-col justify-between transition-all active:scale-95 shadow-xs ${
                    isSelectedJar 
                      ? theme.jarSelected 
                      : 'border-slate-200 bg-white hover:border-slate-300 hover:shadow-xs'
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
                    {isSelectedJar && (
                      <div className={`w-4 h-4 rounded-full flex items-center justify-center text-[10px] ${theme.jarCheck}`}>
                        <Check size={12} className="font-bold" />
                      </div>
                    )}
                  </div>
                  <div className="space-y-1">
                    <div className="flex justify-between text-[10px] font-bold">
                      <span className={`font-mono ${isSelectedJar ? theme.jarBalanceText : 'text-slate-700'}`}>
                        {formatMoney(jar.balance)}
                      </span>
                      {jar.goal_amount > 0 && (
                        <span className="text-slate-400 font-mono">
                          {Math.round(progressPct)}%
                        </span>
                      )}
                    </div>
                    <div className="w-full h-1.5 bg-slate-100 rounded-full overflow-hidden">
                      <div 
                        className={`h-full rounded-full transition-all duration-300 ${isSelectedJar ? theme.jarProgress : 'bg-slate-300'}`} 
                        style={{ width: `${progressPct}%` }}
                      />
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </section>
      </div>

      {/* Bottom Button */}
      <div className="absolute bottom-0 inset-x-0 bg-white/95 backdrop-blur-md border-t border-slate-100 px-5 pt-3 pb-6 z-30 flex flex-col items-center">
        <button 
          onClick={onNext}
          className={`tap-target w-full py-3.5 px-5 text-white font-extrabold text-sm rounded-2xl flex items-center justify-between transition-all duration-200 group active:scale-[0.98] ${theme.ctaGradient}`}
        >
          <div className="flex items-center gap-2.5 text-left">
            <div className="w-8 h-8 rounded-xl bg-white/20 flex items-center justify-center backdrop-blur-sm">
              <CheckCircle size={18} />
            </div>
            <div>
              <span className={`block text-[10px] font-bold uppercase tracking-wider leading-tight ${theme.ctaSubtext}`}>
                Confirm Details
              </span>
              <span className="block text-sm font-black tracking-tight leading-tight">
                {formData.type === 'expense'
                  ? `${theme.actionVerb} ${formatMoney(currentAmount)} • ${formData.fromAccountName || 'Account'}`
                  : formData.type === 'income'
                  ? `${theme.actionVerb} ${formatMoney(currentAmount)} • ${formData.toAccountName || 'Account'}`
                  : `${theme.actionVerb} ${formatMoney(currentAmount)} • ${formData.fromAccountName || 'From'} → ${formData.toAccountName || 'To'}`}
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
