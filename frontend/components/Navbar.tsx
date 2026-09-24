"use client";
import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";
export function Navbar() {
  const pathname = usePathname();
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
          <span className="font-bold text-lg tracking-tight text-neutral-900">
            Paisa
          </span>
        </Link>
      </div>
    </header>
  );
}
