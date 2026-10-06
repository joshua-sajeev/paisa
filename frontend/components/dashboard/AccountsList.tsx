"use client";

import { useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { DashboardAccount } from "@/lib/dashboard";
import { formatMoney } from "@/lib/utils";
import { usePrivacy } from "@/context/PrivacyContext";
import {
  AccountBalanceFill,
  ChevronRight,
} from "@material-symbols-svg/react/w400";

const AVAILABLE_ICONS = ["fbi", "hdfc", "sbi"];

const ACCOUNT_COLORS = [
  "#4F46E5",
  "#059669",
  "#F59E0B",
];

interface AccountsListProps {
  accounts: DashboardAccount[];
}

export default function AccountsList({ accounts }: AccountsListProps) {
  const { isPrivate } = usePrivacy();
  const [revealedIds, setRevealedIds] = useState<Set<string>>(new Set());

  const sortedAccounts = [...accounts].sort((a, b) => {
    if (a.is_primary && !b.is_primary) return -1;
    if (!a.is_primary && b.is_primary) return 1;
    return 0;
  });

  const toggleReveal = (id: string) => {
    if (!isPrivate) return;

    setRevealedIds((prev) => {
      const next = new Set(prev);

      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }

      return next;
    });
  };

  const formatBalance = (amount: number, accountId: string) => {
    const isRevealed = revealedIds.has(accountId);

    if (isPrivate && !isRevealed) {
      return (
        <span className="inline-flex w-fit items-center rounded-md border border-slate-200/70 bg-slate-100/70 px-1.5 py-0.5 text-slate-400 backdrop-blur-sm">
          ₹••••
        </span>
      );
    }

    return <>₹{formatMoney(amount)}</>;
  };

  const getIcon = (iconKey: string, name: string, index: number) => {
    const key = iconKey?.toLowerCase().trim();
    const bankName = name?.toLowerCase().trim();

    const matchedIcon = AVAILABLE_ICONS.find(
      (icon) => key === icon || bankName.includes(icon),
    );

    if (matchedIcon) {
      return (
        <div className="flex h-7 w-7 shrink-0 items-center justify-center overflow-hidden rounded-full bg-white p-1 shadow-sm">
          <Image
            src={`/icons/accounts/${matchedIcon}.svg`}
            alt={name}
            width={24}
            height={24}
            className="h-full w-full object-contain"
          />
        </div>
      );
    }

    const bgColor = ACCOUNT_COLORS[index % ACCOUNT_COLORS.length];

    return (
      <div
        className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full"
        style={{ backgroundColor: bgColor }}
      >
        <AccountBalanceFill
          size={15}
          color="#FFFFFF"
          className="block"
        />
      </div>
    );
  };

  if (!accounts || accounts.length === 0) return null;

  return (
    <section className="flex flex-col gap-2">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h2 className="text-[15px] font-bold tracking-tight text-on-surface">
          Active Accounts
        </h2>

        <Link
          href="/accounts"
          className="flex items-center gap-0.5 text-xs font-semibold text-primary hover:underline"
        >
          View All ({accounts.length})
          <ChevronRight size={14} color="currentColor" />
        </Link>
      </div>

      {/* Horizontal cards */}
      <div
        className="-mx-margin flex gap-2.5 overflow-x-auto px-margin pb-1"
        style={{
          scrollbarWidth: "none",
          msOverflowStyle: "none",
        }}
      >
        {sortedAccounts.map((account, index) => (
          <button
            key={account.id}
            type="button"
            onClick={() => toggleReveal(account.id)}
            className="flex flex-shrink-0 flex-col justify-between rounded-xl bg-white p-3 text-left shadow-sm transition-transform active:scale-[0.98]"
            style={{ width: accounts.length === 1 ? "100%" : "190px" }}
            aria-label={
              isPrivate
                ? `Toggle balance visibility for ${account.name}`
                : `${account.name} account`
            }
          >
            {/* Account info */}
            <div className="flex items-center gap-2">
              {getIcon(account.icon_key, account.name, index)}

              <div className="flex min-w-0 flex-col">
                <span className="truncate text-[13px] font-bold text-on-surface">
                  {account.name}
                </span>

                <span className="truncate text-[10px] text-on-surface-variant">
                  {account.is_primary ? "Primary Account" : "Account"}
                </span>
              </div>
            </div>

            {/* Balance */}
            <div className="flex flex-col pt-2">
              <span className="text-[9px] font-medium uppercase tracking-wider text-gray-500">
                AVAILABLE
              </span>

              <span className="mt-0.5 text-[15px] font-bold tracking-tight text-on-surface">
                {formatBalance(account.balance, account.id)}
              </span>
            </div>
          </button>
        ))}
      </div>
    </section>
  );
}
