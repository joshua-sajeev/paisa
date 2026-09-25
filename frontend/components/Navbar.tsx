"use client";
import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { usePrivacy } from "@/context/PrivacyContext";

export function Navbar() {
  const pathname = usePathname();
  const { isPrivate, togglePrivacy } = usePrivacy();

  if (pathname === "/login") {
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

        {/* Privacy Toggle Button */}
        <button
          type="button"
          onClick={togglePrivacy}
          aria-label="Toggle balances visibility"
          className="w-9 h-9 rounded-full bg-neutral-100 hover:bg-neutral-200 flex items-center justify-center text-neutral-600 transition-all cursor-pointer border border-neutral-200"
        >
          {isPrivate ? (
            /* Closed Eye SVG */
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M9.88 9.88a3 3 0 1 0 4.24 4.24" />
              <path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68" />
              <path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61" />
              <line x1="2" x2="22" y1="2" y2="22" />
            </svg>
          ) : (
            /* Open Eye SVG */
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z" />
              <circle cx="12" cy="12" r="3" />
            </svg>
          )}
        </button>
      </div>
    </header>
  );
}
