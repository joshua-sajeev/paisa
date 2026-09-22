"use client";

import { useEffect } from "react";
import type { Dispatch, SetStateAction } from "react";

const keys = [
  "1",
  "2",
  "3",
  "4",
  "5",
  "6",
  "7",
  "8",
  "9",
  ".",
  "0",
  "backspace",
];

const pinColors = [
  "#ff6700",
  "#ffb7f1",
  "#FFB800",
  "#36dbd5",
  "#002fa7",
];

type PinKeypadProps = {
  pin: string;
  setPinAction: Dispatch<SetStateAction<string>>;
  disabledKeys?: string[];
  maxLength?: number;
  disabled?: boolean;
  shake?: boolean;
  onCompleteAction?: () => void;
};

export function PinKeypad({
  pin,
  setPinAction,
  disabledKeys = [],
  maxLength = 6,
  disabled = false,
  shake = false,
  onCompleteAction,
}: PinKeypadProps) {
  useEffect(() => {
    if (pin.length === maxLength && !disabled) {
      onCompleteAction?.();
    }
  }, [pin, maxLength, disabled, onCompleteAction]);

  function handleKeyPress(key: string) {
    if (disabled) {
      return;
    }

    if (key === "backspace") {
      setPinAction((current) => current.slice(0, -1));
      return;
    }

    if (key === ".") {
      return;
    }

    setPinAction((current) => {
      if (current.length >= maxLength) {
        return current;
      }

      return current + key;
    });
  }

  return (
    <div
      className={`flex w-full flex-col items-center ${shake ? "animate-shake" : ""
        }`}
    >
      {/* PIN status */}
      <div
        aria-label={`${pin.length} of ${maxLength} PIN digits entered`}
        className="mb-5 flex items-center justify-center gap-3.5"
        role="status"
      >
        {Array.from({ length: maxLength }).map((_, index) => {
          const filled = index < pin.length;
          const active = index === pin.length;

          if (filled) {
            const color = pinColors[index % pinColors.length];

            return (
              <div
                key={index}
                className="flex h-5 w-5 scale-110 items-center justify-center rounded-full transition-all duration-200"
                style={{
                  backgroundColor: color,
                  outline: `4px solid ${color}33`,
                  boxShadow: `0 8px 24px -4px ${color}73`,
                }}
              >
                <span className="h-2 w-2 rounded-full bg-white/90" />
              </div>
            );
          }

          if (active) {
            return (
              <div
                key={index}
                className="flex h-5 w-5 items-center justify-center rounded-full border-2 border-dashed border-[#6C47FF]/60 bg-white/80 shadow-inner transition-all duration-200"
              >
                <span className="h-1.5 w-1.5 rounded-full bg-[#6C47FF]/40" />
              </div>
            );
          }

          return (
            <div
              key={index}
              className="h-5 w-5 rounded-full border-2 border-slate-200 bg-white/60 shadow-sm transition-all duration-200"
            />
          );
        })}
      </div>

      {/* Keypad */}
      <div className="mx-auto w-full max-w-[340px]">
        <div className="grid grid-cols-3 justify-items-center gap-x-4 gap-y-3.5">
          {keys.map((key) => {
            const isDisabled =
              disabled ||
              disabledKeys.includes(key) ||
              key === ".";

            return (
              <button
                key={key}
                type="button"
                disabled={isDisabled}
                aria-label={
                  key === "backspace"
                    ? "Delete last digit"
                    : key === "."
                      ? "Decimal point disabled"
                      : `Enter ${key}`
                }
                onClick={() => {
                  handleKeyPress(key);

                  if (navigator.vibrate) {
                    navigator.vibrate(12);
                  }
                }}
                className={
                  key === "backspace"
                    ? "key-pop flex h-[64px] w-[84px] items-center justify-center rounded-[22px] border border-red-100/80 bg-red-50/80 text-[#FF5C67] shadow-[0_4px_0_0_rgba(0,0,0,0.06),0_2px_10px_rgba(108,71,255,0.05)] hover:bg-red-100/80 active:bg-red-200/80 disabled:cursor-not-allowed disabled:opacity-40"
                    : "key-pop flex h-[64px] w-[84px] items-center justify-center rounded-[22px] border border-white/80 bg-white/90 text-[26px] font-bold text-[#111322] shadow-[0_4px_0_0_rgba(0,0,0,0.06),0_2px_10px_rgba(108,71,255,0.05)] hover:bg-white active:bg-[#F3EEFF] disabled:cursor-not-allowed disabled:opacity-40"
                }
              >
                {key === "backspace" ? (
                  <svg
                    className="h-6 w-6"
                    fill="none"
                    stroke="currentColor"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2.2"
                    viewBox="0 0 24 24"
                  >
                    <path d="M21 4H8l-7 8 7 8h13a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2z" />
                    <line x1="18" y1="9" x2="12" y2="15" />
                    <line x1="12" y1="9" x2="18" y2="15" />
                  </svg>
                ) : key === "." ? (
                  <span className="leading-none text-slate-400">
                    •
                  </span>
                ) : (
                  <span className="leading-none">{key}</span>
                )}
              </button>
            );
          })}
        </div>
      </div>
    </div>
  );
}
