"use client";

import { useRef, useState } from "react";
import { useRouter } from "next/navigation";
import {
  formatMoney,
  formatTime,
  getTypeClasses,
  TransactionIcon,
} from "./transaction-utils";

import type { Transaction } from "./transaction-utils";
import Link from "next/link";

const labelColors = ["#6366F1", "#07B682", "#F59E0B", "#EC4899", "#8B5CF6"];

function getLabelColor(label: string): string {
  let hash = 0;
  for (let index = 0; index < label.length; index++) {
    hash = (hash * 31 + label.charCodeAt(index)) >>> 0;
  }
  return labelColors[hash % labelColors.length];
}

type Props = {
  transaction: Transaction;
  onDelete: (transaction: Transaction) => void;
};

function EditIcon() {
  return <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d="M12 20h9"/><path d="M16.5 3.5a2.12 2.12 0 1 1 3 3L7 19l-4 1 1-4 12.5-12.5z"/></svg>;
}

function DeleteIcon() {
  return <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/><path d="M10 11v6"/><path d="M14 11v6"/><path d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/></svg>;
}

export default function TransactionCard({ transaction, onDelete }: Props) {
  const router = useRouter();
  const [offset, setOffset] = useState(0);
  const [dragging, setDragging] = useState(false);
  const startX = useRef(0);
  const startOffset = useRef(0);
  const currentOffset = useRef(0);
  const didDrag = useRef(false);
  const actionWidth = 144;

  const styles = getTypeClasses(transaction.type);

  const amountPrefix =
    transaction.type === "income"
      ? "+"
      : transaction.type === "expense"
        ? "-"
        : "";

  const category = transaction.category
    ? transaction.category.charAt(0).toUpperCase() + transaction.category.slice(1)
    : "Uncategorized";
  const accountColor = getLabelColor(transaction.account);

  function handlePointerDown(event: React.PointerEvent<HTMLDivElement>) {
    startX.current = event.clientX;
    startOffset.current = offset;
    currentOffset.current = offset;
    didDrag.current = false;
    setDragging(true);
    event.currentTarget.setPointerCapture(event.pointerId);
  }

  function handlePointerMove(event: React.PointerEvent<HTMLDivElement>) {
    if (!dragging) return;
    const delta = event.clientX - startX.current;
    if (Math.abs(delta) > 4) didDrag.current = true;
    const next = Math.max(-actionWidth, Math.min(0, startOffset.current + delta));
    currentOffset.current = next;
    setOffset(next);
  }

  function handlePointerUp() {
    setDragging(false);
    const next = currentOffset.current <= -60 ? -actionWidth : 0;
    currentOffset.current = next;
    setOffset(next);
  }

  function handleClickCapture(event: React.MouseEvent<HTMLAnchorElement>) {
    if (didDrag.current) {
      event.preventDefault();
      didDrag.current = false;
    }
  }

  return (
    <div className="relative overflow-hidden">
      <div className="absolute inset-y-0 right-0 flex">
        <button type="button" onClick={() => { setOffset(0); router.push(`/transactions/edit?id=${encodeURIComponent(transaction.id)}`); }} className="flex w-[72px] items-center justify-center bg-white text-[#067FF3]" aria-label="Edit transaction"><EditIcon/></button>
        <button type="button" onClick={() => { setOffset(0); onDelete(transaction); }} className="flex w-[72px] items-center justify-center bg-white text-[#EF4444]" aria-label="Delete transaction"><DeleteIcon/></button>
      </div>

      <div
        className="relative flex select-none items-center gap-3 bg-white px-3.5 py-3 touch-pan-y"
        style={{ transform: `translateX(${offset}px)`, transition: dragging ? "none" : "transform 180ms ease-out" }}
        onPointerDown={handlePointerDown}
        onPointerMove={handlePointerMove}
        onPointerUp={handlePointerUp}
        onPointerCancel={handlePointerUp}
      >
        <div className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-xl ${styles.icon}`}>
          <TransactionIcon type={transaction.type} category={transaction.category}/>
        </div>

        <Link href={`/transactions/edit?id=${encodeURIComponent(transaction.id)}`} onClickCapture={handleClickCapture} className="min-w-0 flex-1">
          <div className="flex items-center justify-between gap-2">
            <div className="flex min-w-0 items-center gap-1.5">
              <h3 className="truncate text-[14px] font-bold leading-tight text-[#0b1c30]">{transaction.name}</h3>
              <span className="shrink-0 rounded-md border px-1.5 py-0.5 text-[11px] font-semibold" style={{ color: accountColor, backgroundColor: `${accountColor}1A`, borderColor: `${accountColor}33` }}>{transaction.account}</span>
            </div>
            <p className={`shrink-0 text-sm font-bold tracking-tight ${styles.amount}`}>{amountPrefix}{formatMoney(transaction.amount)}</p>
          </div>
          <div className="mt-0.5 flex min-w-0 items-center justify-between gap-2 text-[11px] text-[#737686]"><p className="min-w-0 truncate">{category}{transaction.jar_name && <span> • {transaction.jar_name}</span>}</p><span className="ml-auto shrink-0 text-right text-[#737686]/70">{formatTime(transaction.occurred_at)}</span></div>
        </Link>
      </div>
    </div>
  );
}
