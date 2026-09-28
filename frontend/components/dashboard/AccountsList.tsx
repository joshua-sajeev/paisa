'use client';

import Image from 'next/image';
import { DashboardAccount } from '@/lib/dashboard';
import { usePrivacy } from '@/context/PrivacyContext';
import {
  AccountBalanceFill,
  ChevronRight,
} from '@material-symbols-svg/react/w400';

const AVAILABLE_ICONS = ['fbi', 'hdfc', 'sbi'];

interface AccountsListProps {
  accounts: DashboardAccount[];
}

export default function AccountsList({ accounts }: AccountsListProps) {
  const { isPrivate } = usePrivacy();

  const sortedAccounts = [...accounts].sort((a, b) => {
    if (a.is_primary && !b.is_primary) return -1;
    if (!a.is_primary && b.is_primary) return 1;
    return 0;
  });

  const formatBalance = (amount: number) => {
    if (isPrivate) return '₹••••••••';

    return new Intl.NumberFormat('en-IN', {
      style: 'currency',
      currency: 'INR',
      maximumFractionDigits: 0,
    }).format(amount);
  };
  const ACCOUNT_COLORS = [
    '#4F46E5',
    '#059669',
    '#F59E0B',
  ];
  const getIcon = (iconKey: string, name: string, index: number) => {
    const key = iconKey?.toLowerCase().trim();
    const bankName = name?.toLowerCase().trim();

    const matchedIcon = AVAILABLE_ICONS.find(
      (icon) => key === icon || bankName.includes(icon)
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
        <h2 className="text-[15px] font-bold text-on-surface tracking-tight">
          Active Accounts
        </h2>

        <button
          type="button"
          className="text-xs font-semibold text-primary flex items-center gap-0.5 hover:underline"
        >
          View All ({accounts.length})
          <ChevronRight width={14} height={14} />
        </button>
      </div>

      {/* Horizontal cards */}
      <div
        className="flex gap-2.5 overflow-x-auto pb-1 -mx-margin px-margin"
        style={{
          scrollbarWidth: 'none',
          msOverflowStyle: 'none',
        }}
      >
        {sortedAccounts.map((account, index) => (
          <div
            key={account.id}
            className="flex flex-col justify-between flex-shrink-0 bg-white rounded-xl p-3 shadow-sm"
            style={{ width: '190px' }}
          >
            {/* Account info */}
            <div className="flex items-center gap-2">
              {getIcon(account.icon_key, account.name, index)}

              <div className="flex flex-col min-w-0">
                <span className="text-[13px] font-bold text-on-surface truncate">
                  {account.name}
                </span>

                <span className="text-[10px] text-on-surface-variant truncate">
                  {account.is_primary ? 'Primary Account' : 'Account'}
                </span>
              </div>
            </div>

            {/* Balance */}
            <div className="flex flex-col pt-2">
              <span className="text-[9px] uppercase tracking-wider text-gray-500 font-medium">
                AVAILABLE
              </span>

              <span className="text-[15px] font-bold text-on-surface tracking-tight mt-0.5">
                {formatBalance(account.balance)}
              </span>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
