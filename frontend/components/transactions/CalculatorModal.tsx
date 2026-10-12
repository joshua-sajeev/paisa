"use client";

import { useEffect, useMemo, useState } from "react";
import { Close, Backspace, ArrowForward, Calculate } from "@material-symbols-svg/react/w400";
import {
  appendDecimal,
  appendDigit,
  appendOperator,
  backspace,
  evaluate,
  formatExpression,
  formatIndian,
  isOperator,
  toAmountString,
} from "@/lib/calculator";

export type CalculatorAccent = {
  /** hex colour, used for the ₹ symbol */
  color: string;
  bgClass: string;
  hoverClass: string;
  shadowClass: string;
};

type Props = {
  /** Current amount from the keypad; used to seed the calculator. */
  initialAmount: string;
  accent: CalculatorAccent;
  onClose: () => void;
  /** Receives the amount as a string, e.g. "1630" or "12.5". */
  onApply: (amount: string) => void;
};

const ERROR_TEXT = {
  "divide-by-zero": "Can't divide by zero",
  "too-large": "Amount is too large",
} as const;

const keyBase =
  "h-11 rounded-xl font-semibold active:scale-95 transition-all select-none";
const digitKey = `${keyBase} hover:bg-slate-100 text-slate-800 text-lg`;
const opKey = `${keyBase} bg-slate-100 hover:bg-slate-200 text-slate-700 text-base`;

export default function CalculatorModal({ initialAmount, accent, onClose, onApply }: Props) {
  // Mounted only while open, so this seeds fresh every time the modal opens.
  const [expr, setExpr] = useState(() => {
    const n = parseFloat(initialAmount);
    return n > 0 ? initialAmount : "";
  });
  const [shown, setShown] = useState(false);

  const result = useMemo(() => evaluate(expr), [expr]);
  const canApply = result.ok && result.value > 0;
  const hasOperator = useMemo(
    () => [...expr].some((c, i) => isOperator(c) && !(i === 0 && c === "-")),
    [expr]
  );

  // slide-up on mount
  useEffect(() => {
    const id = requestAnimationFrame(() => setShown(true));
    return () => cancelAnimationFrame(id);
  }, []);

  const handleEquals = () => {
    setExpr((prev) => {
      const r = evaluate(prev);
      return r.ok ? toAmountString(r.value) : prev;
    });
  };

  const handleApply = () => {
    if (result.ok && result.value > 0) onApply(toAmountString(result.value));
  };

  // Keyboard support: digits, . + - * /, Enter / =, Backspace, Delete / c, Esc
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.ctrlKey || e.metaKey || e.altKey) return;
      const k = e.key;

      if (/^\d$/.test(k)) setExpr((p) => appendDigit(p, k));
      else if (k === ".") setExpr((p) => appendDecimal(p));
      else if (k === "+" || k === "-" || k === "*" || k === "/") {
        e.preventDefault(); // "/" opens quick-find in some browsers
        setExpr((p) => appendOperator(p, k));
      } else if (k === "x" || k === "X") setExpr((p) => appendOperator(p, "*"));
      else if (k === "Enter" || k === "=") {
        e.preventDefault();
        handleEquals();
      } else if (k === "Backspace") setExpr((p) => backspace(p));
      else if (k === "Delete" || k === "c" || k === "C") setExpr("");
      else if (k === "Escape") onClose();
      else return;
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  const resultText = result.ok ? formatIndian(toAmountString(result.value)) : "0";

  return (
    <div
      className={`fixed inset-0 z-50 flex flex-col justify-end bg-slate-900/40 backdrop-blur-[2px] select-none transition-opacity duration-300 ${
        shown ? "opacity-100" : "opacity-0"
      }`}
      onClick={onClose}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-label="Quick calculator"
        onClick={(e) => e.stopPropagation()}
        className={`w-full max-w-[480px] mx-auto bg-white rounded-t-[28px] shadow-2xl border-t border-slate-100 flex flex-col overflow-hidden max-h-[88%] transition-transform duration-300 ease-out ${
          shown ? "translate-y-0" : "translate-y-full"
        }`}
      >
        {/* Grab handle */}
        <div className="w-full pt-3 pb-1 flex justify-center">
          <div className="w-10 h-1 bg-slate-200 rounded-full" />
        </div>

        {/* Header */}
        <div className="px-5 py-2 flex items-center justify-between border-b border-slate-100">
          <div className="flex items-center gap-2">
            <div className="w-7 h-7 rounded-lg bg-slate-100 flex items-center justify-center text-slate-600">
              <Calculate size={16} />
            </div>
            <div>
              <h3 className="text-xs font-bold text-slate-800 uppercase tracking-wide">
                Quick Calculator
              </h3>
              <p className="text-[10px] text-slate-400">Sum multiple expenses or splits</p>
            </div>
          </div>
          <button
            type="button"
            aria-label="Close calculator"
            onClick={onClose}
            className="w-8 h-8 rounded-full flex items-center justify-center text-slate-400 hover:text-slate-700 hover:bg-slate-100 transition-colors"
          >
            <Close size={20} />
          </button>
        </div>

        {/* Display */}
        <div className="px-5 py-3 bg-slate-50/70 flex flex-col justify-center items-end border-b border-slate-100">
          <div className="text-xs font-medium text-slate-400 tracking-wide font-mono h-4 overflow-hidden whitespace-nowrap text-right w-full">
            {expr ? formatExpression(expr) : "Type an expression"}
          </div>
          <div className="flex items-baseline gap-1 text-slate-900 font-extrabold">
            <span className="text-lg text-slate-400">=</span>
            <span className="text-xs font-bold" style={{ color: accent.color }}>
              ₹
            </span>
            <span className="text-3xl font-extrabold tracking-tight font-mono">{resultText}</span>
          </div>
          <div className="h-4 text-[11px] font-medium text-red-500">
            {!result.ok && result.error !== "incomplete" ? ERROR_TEXT[result.error] : ""}
          </div>
        </div>

        {/* Keys */}
        <div className="p-4 grid grid-cols-4 gap-2 bg-white">
          <button type="button" onClick={() => setExpr("")} className={`${opKey} text-red-500 font-bold text-sm`}>
            C
          </button>
          <button type="button" onClick={() => setExpr((p) => appendOperator(p, "/"))} className={opKey}>
            ÷
          </button>
          <button type="button" onClick={() => setExpr((p) => appendOperator(p, "*"))} className={opKey}>
            ×
          </button>
          <button
            type="button"
            aria-label="Delete last"
            onClick={() => setExpr((p) => backspace(p))}
            className={`${opKey} flex items-center justify-center`}
          >
            <Backspace size={18} />
          </button>

          {["7", "8", "9"].map((d) => (
            <button type="button" key={d} onClick={() => setExpr((p) => appendDigit(p, d))} className={digitKey}>
              {d}
            </button>
          ))}
          <button type="button" onClick={() => setExpr((p) => appendOperator(p, "-"))} className={opKey}>
            −
          </button>

          {["4", "5", "6"].map((d) => (
            <button type="button" key={d} onClick={() => setExpr((p) => appendDigit(p, d))} className={digitKey}>
              {d}
            </button>
          ))}
          <button type="button" onClick={() => setExpr((p) => appendOperator(p, "+"))} className={opKey}>
            +
          </button>

          {["1", "2", "3"].map((d) => (
            <button type="button" key={d} onClick={() => setExpr((p) => appendDigit(p, d))} className={digitKey}>
              {d}
            </button>
          ))}
          <button
            type="button"
            onClick={handleEquals}
            disabled={!hasOperator}
            className="row-span-2 h-full rounded-xl bg-slate-800 hover:bg-slate-900 text-white font-bold text-lg active:scale-95 transition-all flex items-center justify-center disabled:opacity-40 disabled:active:scale-100"
          >
            =
          </button>

          <button type="button" onClick={() => setExpr((p) => appendDigit(p, "0"))} className={`${digitKey} col-span-2`}>
            0
          </button>
          <button type="button" onClick={() => setExpr((p) => appendDecimal(p))} className={`${digitKey} font-bold`}>
            •
          </button>
        </div>

        {/* Actions */}
        <div className="px-4 pt-1 flex items-center gap-2 pb-[max(1rem,env(safe-area-inset-bottom))]">
          <button
            type="button"
            onClick={onClose}
            className="flex-1 py-3 px-4 rounded-xl border border-slate-200 text-xs font-semibold text-slate-600 hover:bg-slate-50 active:scale-95 transition-all"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={handleApply}
            disabled={!canApply}
            className={`flex-[2] py-3 px-4 rounded-xl font-bold text-xs text-white shadow-md flex items-center justify-center gap-1.5 transition-all duration-200 active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed disabled:active:scale-100 ${accent.bgClass} ${accent.hoverClass} ${accent.shadowClass}`}
          >
            <span>{canApply ? `Apply ₹${resultText} to Amount` : "Apply to Amount"}</span>
            <ArrowForward size={14} />
          </button>
        </div>
      </div>
    </div>
  );
}
