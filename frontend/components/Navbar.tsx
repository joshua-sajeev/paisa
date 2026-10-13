"use client";
import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import { usePrivacy } from "@/context/PrivacyContext";
import { ApiError, apiFetch } from "@/lib/api";
import {
  RestartAlt,
  Visibility,
  VisibilityOff,
} from '@material-symbols-svg/react/w400';

export function Navbar() {
  const pathname = usePathname();
  const { isPrivate, togglePrivacy } = usePrivacy();
  const [demoMode, setDemoMode] = useState(false);
  const [resettingDemo, setResettingDemo] = useState(false);
  const [showResetConfirmation, setShowResetConfirmation] = useState(false);
  const [resetError, setResetError] = useState<string | null>(null);

  useEffect(() => {
    if (pathname === "/login" || pathname === "/transactions/add" || pathname.endsWith("/edit")) return;

    let active = true;
    apiFetch<{ enabled: boolean }>("/api/v1/demo/mode")
      .then(({ enabled }) => {
        if (active) setDemoMode(enabled);
      })
      .catch(() => {
        if (active) setDemoMode(false);
      });

    return () => {
      active = false;
    };
  }, [pathname]);

  const resetDemo = async () => {
    setResettingDemo(true);
    setResetError(null);
    try {
      await apiFetch<{ message: string }>("/api/v1/demo/reset", {
        method: "POST",
      });
      window.location.reload();
    } catch (error) {
      const message = error instanceof ApiError
        ? error.message
        : "Could not reset demo data. Please try again.";
      setResetError(message);
    } finally {
      setResettingDemo(false);
    }
  };

  if (pathname === "/login" || pathname === "/transactions/add" || pathname.endsWith("/edit")) {
    return null;
  }

  return (
    <header className="sticky top-0 z-50 bg-[#FFFAF0] px-6 py-3">
      <div className="flex items-center justify-between max-w-7xl mx-auto">
        <Link
          href="/dashboard"
          className="flex items-center gap-2.5 hover:opacity-80 transition-opacity"
        >
          <Image
            src="/logo.svg"
            alt="Paisa Logo"
            width={32}
            height={32}
            priority
          />
          <span className="font-bold text-sm tracking-tight text-neutral-900">
            PAISA
          </span>
        </Link>

        <div className="flex items-center gap-2">
          {demoMode && (
            <button
              onClick={() => {
                setResetError(null);
                setShowResetConfirmation(true);
              }}
              disabled={resettingDemo}
              className="inline-flex h-9 items-center gap-1.5 rounded-full border border-neutral-200 bg-neutral-100 px-3 text-xs font-semibold text-neutral-700 transition-colors hover:bg-neutral-200 disabled:cursor-wait disabled:opacity-60"
            >
              <RestartAlt width={17} height={17} />
              {resettingDemo ? "Resetting…" : "Reset demo"}
            </button>
          )}

          {pathname === "/dashboard" && (
            <button
              onClick={togglePrivacy}
              aria-label="Toggle balances visibility"
              className="w-9 h-9 rounded-full bg-neutral-100 hover:bg-neutral-200 flex items-center justify-center text-neutral-600 transition-all cursor-pointer border border-neutral-200"
            >
              {isPrivate ? (
                <Visibility width={20} height={20} />
              ) : (
                <VisibilityOff width={20} height={20} />
              )}
            </button>
          )}
        </div>
      </div>
      {showResetConfirmation && (
        <div
          className="fixed inset-0 z-[100] flex items-center justify-center bg-neutral-950/40 px-4 backdrop-blur-sm"
          onClick={() => {
            if (!resettingDemo) setShowResetConfirmation(false);
          }}
        >
          <section
            aria-labelledby="reset-demo-title"
            aria-describedby="reset-demo-description"
            aria-modal="true"
            role="dialog"
            className="w-full max-w-md rounded-3xl border border-amber-100 bg-[#FFFAF0] p-6 shadow-2xl shadow-neutral-950/20"
            onClick={(event) => event.stopPropagation()}
          >
            <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-amber-100 text-amber-800">
              <RestartAlt width={24} height={24} />
            </div>
            <h2 id="reset-demo-title" className="mt-4 text-lg font-bold text-neutral-900">
              Reset demo data?
            </h2>
            <p id="reset-demo-description" className="mt-2 text-sm leading-6 text-neutral-600">
              This replaces your current demo accounts, jars, goals, and transactions with the sample data.
            </p>
            {resetError && (
              <p role="alert" className="mt-4 rounded-xl border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-700">
                {resetError}
              </p>
            )}
            <div className="mt-6 flex justify-end gap-2">
              <button
                type="button"
                disabled={resettingDemo}
                onClick={() => setShowResetConfirmation(false)}
                className="rounded-xl px-4 py-2.5 text-sm font-semibold text-neutral-600 transition hover:bg-white disabled:opacity-50"
              >
                Cancel
              </button>
              <button
                type="button"
                disabled={resettingDemo}
                onClick={resetDemo}
                className="inline-flex items-center gap-2 rounded-xl bg-amber-800 px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-amber-900 disabled:cursor-wait disabled:opacity-60"
              >
                <RestartAlt width={17} height={17} />
                {resettingDemo ? "Resetting…" : "Reset demo"}
              </button>
            </div>
          </section>
        </div>
      )}
    </header>
  );
}
