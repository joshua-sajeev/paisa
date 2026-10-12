'use client';

import { useRef, useState } from 'react';
import { formatMoney } from '@/lib/utils';
import { DashboardTransaction } from '@/lib/dashboard';
import { getTypeClasses, TransactionIcon } from '@/components/transactions/transaction-utils';


interface TransactionRowProps {
  transaction: DashboardTransaction;
  onEdit?: (transaction: DashboardTransaction) => void;
  onDelete?: (transaction: DashboardTransaction) => void;
}

function formatDate(value: string) {
  const date = new Date(value);
  if (!Number.isFinite(date.getTime())) return '—';

  return date.toLocaleDateString('en-IN', {
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

export default function TransactionRow({
  transaction,
  onEdit,
  onDelete,
}: TransactionRowProps) {
  const [offset, setOffset] = useState(0);
  const [dragging, setDragging] = useState(false);

  const startX = useRef(0);
  const startOffset = useRef(0);

  const isTransfer = transaction.type === 'transfer';
  const isIncome = transaction.type === 'income';
  const styles = getTypeClasses(transaction.type);

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
          className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-xl ${styles.icon}`}
        >
          <TransactionIcon type={transaction.type} category={transaction.category} />
        </div>

        {/* Transaction information */}
        <div className="flex min-w-0 flex-1 flex-col">
          <span className="truncate text-[14px] font-bold leading-tight text-slate-900">
            {transaction.name}
          </span>

          <div className="mt-1 flex items-center gap-1 text-[11px] font-medium leading-tight text-slate-500">
            <span>{categoryLabel}</span>
            <span>•</span>
            <span>{formatDate(transaction.occurred_at)}</span>
          </div>
        </div>

        {/* Amount + Jar + Bank */}
        <div className="ml-2 flex shrink-0 flex-col items-end text-right">
          {/* Amount */}
          <span
            className={`text-[14px] font-bold leading-tight ${styles.amount}`}
          >
            {displayAmount}
          </span>

          {/* Jar + Bank */}
          <div className="mt-1 max-w-[150px] truncate">
            <span className="text-[11px] font-medium leading-tight text-slate-500">
              {accountInfo}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
