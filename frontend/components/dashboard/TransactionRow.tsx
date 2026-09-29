'use client';

import { useRef, useState } from 'react';
import { formatMoney } from '@/lib/utils';
import { DashboardTransaction } from '@/lib/dashboard';
import { usePrivacy } from '@/context/PrivacyContext';

import {
  LocalDining,
  DirectionsCar,
  Movie,
  ShoppingCart,
  MedicalServices,
  SwapHoriz,
  Favorite,
  TrendingUp,
  Home,
  Label,
} from '@material-symbols-svg/react/w400';

interface TransactionRowProps {
  transaction: DashboardTransaction;
  onEdit?: (transaction: DashboardTransaction) => void;
  onDelete?: (transaction: DashboardTransaction) => void;
}

const TRANSACTION_STYLES = {
  food: {
    icon: LocalDining,
    background: '#F7CD1B',
    iconColor: '#171717',
  },
  transport: {
    icon: DirectionsCar,
    background: '#4F46E5',
    iconColor: '#FFFFFF',
  },
  entertainment: {
    icon: Movie,
    background: '#9333EA',
    iconColor: '#FFFFFF',
  },
  groceries: {
    icon: ShoppingCart,
    background: '#16A34A',
    iconColor: '#FFFFFF',
  },
  health: {
    icon: MedicalServices,
    background: '#EC4899',
    iconColor: '#FFFFFF',
  },
  transfer: {
    icon: SwapHoriz,
    background: '#067FF3',
    iconColor: '#FFFFFF',
  },
  donation: {
    icon: Favorite,
    background: '#DC2626',
    iconColor: '#FFFFFF',
  },
  investment: {
    icon: TrendingUp,
    background: '#059669',
    iconColor: '#FFFFFF',
  },
  housing: {
    icon: Home,
    background: '#D97706',
    iconColor: '#FFFFFF',
  },
  other: {
    icon: Label,
    background: '#6B7280',
    iconColor: '#FFFFFF',
  },
} as const;

function formatDate(date: string) {
  return new Date(date).toLocaleDateString('en-IN', {
    month: 'short',
    day: 'numeric',
  });
}

function capitalize(value: string) {
  return value.charAt(0).toUpperCase() + value.slice(1);
}

function EditIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="18"
      height="18"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M12 20h9" />
      <path d="M16.5 3.5a2.12 2.12 0 1 1 3 3L7 19l-4 1 1-4 12.5-12.5z" />
    </svg>
  );
}

function DeleteIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="18"
      height="18"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <polyline points="3 6 5 6 21 6" />
      <path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6" />
      <path d="M10 11v6" />
      <path d="M14 11v6" />
      <path d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2" />
    </svg>
  );
}

function PrivacyPill({
  children,
  size = 'amount',
}: {
  children: React.ReactNode;
  size?: 'amount' | 'account';
}) {
  return (
    <span
      className={
        size === 'amount'
          ? 'inline-flex w-fit items-center rounded-lg border border-slate-200/70 bg-slate-100/70 px-2 py-0.5 text-[12px] font-bold text-slate-400 backdrop-blur-sm'
          : 'inline-flex w-fit items-center rounded-md border border-slate-200/70 bg-slate-100/70 px-1.5 py-0.5 text-[10px] font-medium text-slate-400 backdrop-blur-sm'
      }
    >
      {children}
    </span>
  );
}

export default function TransactionRow({
  transaction,
  onEdit,
  onDelete,
}: TransactionRowProps) {
  const { isPrivate } = usePrivacy();

  const [offset, setOffset] = useState(0);
  const [dragging, setDragging] = useState(false);

  const startX = useRef(0);
  const startOffset = useRef(0);

  const isTransfer = transaction.type === 'transfer';
  const isIncome = transaction.type === 'income';

  const style = isTransfer
    ? TRANSACTION_STYLES.transfer
    : TRANSACTION_STYLES[transaction.category] ??
      TRANSACTION_STYLES.other;

  const Icon = style.icon;

  const ACTION_WIDTH = 144;
  const SWIPE_THRESHOLD = 60;

  const handlePointerDown = (
    event: React.PointerEvent<HTMLDivElement>,
  ) => {
    startX.current = event.clientX;
    startOffset.current = offset;
    setDragging(true);

    event.currentTarget.setPointerCapture(event.pointerId);
  };

  const handlePointerMove = (
    event: React.PointerEvent<HTMLDivElement>,
  ) => {
    if (!dragging) return;

    const delta = event.clientX - startX.current;

    let nextOffset = startOffset.current + delta;

    nextOffset = Math.min(0, nextOffset);
    nextOffset = Math.max(-ACTION_WIDTH, nextOffset);

    setOffset(nextOffset);
  };

  const handlePointerUp = () => {
    setDragging(false);

    if (offset <= -SWIPE_THRESHOLD) {
      setOffset(-ACTION_WIDTH);
    } else {
      setOffset(0);
    }
  };

  const handleEdit = () => {
    setOffset(0);
    onEdit?.(transaction);
  };

  const handleDelete = () => {
    setOffset(0);
    onDelete?.(transaction);
  };

  const displayAmount = isTransfer
    ? `₹${formatMoney(transaction.amount)}`
    : `${isIncome ? '+' : '-'}₹${formatMoney(transaction.amount)}`;

  const amountColor = isTransfer
    ? '#067FF3'
    : isIncome
      ? '#10B981'
      : '#EF4444';

  const categoryLabel = capitalize(transaction.category);

  const accountInfo = transaction.jar_name
    ? `${transaction.jar_name} • ${transaction.account}`
    : transaction.account;

  return (
    <div className="relative overflow-hidden rounded-xl">
      {/* Swipe actions */}
      <div className="absolute inset-y-0 right-0 flex">
        <button
          type="button"
          onClick={handleEdit}
          className="flex w-[72px] cursor-pointer items-center justify-center bg-white text-[#067FF3]"
          aria-label="Edit transaction"
        >
          <EditIcon />
        </button>

        <button
          type="button"
          onClick={handleDelete}
          className="flex w-[72px] cursor-pointer items-center justify-center bg-white text-[#EF4444]"
          aria-label="Delete transaction"
        >
          <DeleteIcon />
        </button>
      </div>

      {/* Main row */}
      <div
        className="relative flex select-none items-center gap-3 bg-white px-3 py-3 touch-pan-y"
        style={{
          transform: `translateX(${offset}px)`,
          transition: dragging
            ? 'none'
            : 'transform 180ms ease-out',
        }}
        onPointerDown={handlePointerDown}
        onPointerMove={handlePointerMove}
        onPointerUp={handlePointerUp}
        onPointerCancel={handlePointerUp}
      >
        {/* Category icon */}
        <div
          className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl"
          style={{
            backgroundColor: style.background,
          }}
        >
          <Icon
            className="h-5 w-5"
            style={{
              color: style.iconColor,
            }}
          />
        </div>

        {/* Transaction information */}
        <div className="flex min-w-0 flex-1 flex-col">
          <span className="truncate text-[14px] font-bold leading-tight text-slate-900">
            {transaction.name}
          </span>

          <div className="mt-1 flex items-center gap-1 text-[11px] font-medium leading-tight text-slate-500">
            <span>{categoryLabel}</span>
            <span>•</span>
            <span>{formatDate(transaction.date)}</span>
          </div>
        </div>

        {/* Amount + Jar + Bank */}
        <div className="ml-2 flex shrink-0 flex-col items-end text-right">
          {/* Amount */}
          {isPrivate ? (
            <PrivacyPill size="amount">₹••••</PrivacyPill>
          ) : (
            <span
              className="text-[14px] font-bold leading-tight"
              style={{
                color: amountColor,
              }}
            >
              {displayAmount}
            </span>
          )}

          {/* Jar + Bank */}
          <div className="mt-1 max-w-[150px] truncate">
            {isPrivate ? (
              <PrivacyPill size="account">••••</PrivacyPill>
            ) : (
              <span className="text-[11px] font-medium leading-tight text-slate-500">
                {accountInfo}
              </span>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
