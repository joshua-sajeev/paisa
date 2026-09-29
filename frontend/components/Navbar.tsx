"use client";
import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { usePrivacy } from "@/context/PrivacyContext";
import {
  Visibility,
  VisibilityOff,
} from '@material-symbols-svg/react/w400';

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
      </div>
    </header>
  );
}
