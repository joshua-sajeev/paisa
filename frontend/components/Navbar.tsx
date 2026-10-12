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

  useEffect(() => {
    if (pathname === "/login" || pathname === "/transactions/add") return;

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
    const confirmed = window.confirm(
      "Reset demo data? This replaces the current accounts, jars, goals, and transactions with the seeded sample data.",
    );
    if (!confirmed) return;

    setResettingDemo(true);
    try {
      await apiFetch<{ message: string }>("/api/v1/demo/reset", {
        method: "POST",
      });
      window.location.reload();
    } catch (error) {
      const message = error instanceof ApiError
        ? error.message
        : "Could not reset demo data. Please try again.";
      window.alert(message);
    } finally {
      setResettingDemo(false);
    }
  };

  if (pathname === "/login" || pathname === "/transactions/add") {
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
              onClick={resetDemo}
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
    </header>
  );
}
