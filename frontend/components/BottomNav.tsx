'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';

import {
  Home,
  HomeFill,
  ReceiptLong,
  ReceiptLongFill,
  AccountBalance,
  AccountBalanceFill,
  Savings,
  SavingsFill,
  AdsClick,
  AdsClickFill,
} from '@material-symbols-svg/react/w400';

const NAV_ITEMS = [
  {
    label: 'Home',
    path: '/dashboard',
    icon: Home,
    activeIcon: HomeFill,
  },
  {
    label: 'Transactions',
    path: '/transactions',
    icon: ReceiptLong,
    activeIcon: ReceiptLongFill,
  },
  {
    label: 'Accounts',
    path: '/accounts',
    icon: AccountBalance,
    activeIcon: AccountBalanceFill,
  },
  {
    label: 'Jars',
    path: '/jars',
    icon: Savings,
    activeIcon: SavingsFill,
  },
  {
    label: 'Goals',
    path: '/goals',
    icon: AdsClick,
    activeIcon: AdsClickFill,
  },
];

export default function BottomNav() {
  const pathname = usePathname();

  return (
    <nav className="fixed bottom-0 z-50 w-full bg-white/85 pb-safe backdrop-blur-xl shadow-[0_-4px_20px_rgba(0,0,0,0.04)]">
      <div className="relative flex h-20 items-center justify-between px-2">
        {NAV_ITEMS.map((item) => {
          const isActive =
            item.path === '/dashboard'
              ? pathname === '/dashboard'
              : pathname.startsWith(item.path);

          const Icon = isActive ? item.activeIcon : item.icon;

          return (
            <Link
              key={item.path}
              href={item.path}
              aria-current={isActive ? 'page' : undefined}
              className={`flex min-h-[48px] min-w-[48px] flex-1 flex-col items-center justify-center transition-colors ${
                isActive
                  ? 'font-bold text-[#067FF3]'
                  : 'text-slate-500 hover:text-slate-900'
              }`}
            >
              <Icon
                size={22}
                color="currentColor"
              />

              <span className="mt-0.5 text-[10px] font-semibold leading-tight">
                {item.label}
              </span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
