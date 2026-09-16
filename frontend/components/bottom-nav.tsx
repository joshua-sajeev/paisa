"use client";

import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";

const navItems = [
  {
    href: "/dashboard",
    label: "Home",
    icon: "home",
  },
  {
    href: "/transactions",
    label: "Transactions",
    icon: "receipt_long",
  },
  {
    href: "/accounts",
    label: "Accounts",
    icon: "account_balance",
  },
  {
    href: "/jars",
    label: "Jars",
    icon: "savings",
  },
  {
    href: "/goals",
    label: "Goals",
    icon: "track_changes",
  },
];

export function BottomNav() {
  const pathname = usePathname();

  return (
    <nav className="fixed bottom-0 left-1/2 z-40 w-full max-w-[430px] -translate-x-1/2 border-t border-slate-100 bg-[#f6f8fd] px-3 py-2 shadow-lg backdrop-blur-md">
      <div className="grid grid-cols-5 items-center text-center">
        {navItems.map((item) => {
          const active = pathname === item.href;

          return (
            <Link
              key={item.href}
              href={item.href}
              aria-current={active ? "page" : undefined}
              className={`group flex flex-col items-center justify-center py-1 transition ${active
                ? "text-indigo-600"
                : "text-slate-400 hover:text-slate-700"
                }`}
            >
              <Image
                src={`/icons/${item.icon}-${active ? "filled" : "outlined"}.svg`}
                alt=""
                width={24}
                height={24}
                className="mb-0.5 transition-transform group-hover:scale-110"
              />

              <span
                className={`text-[10px] tracking-tight ${active ? "font-extrabold" : "font-semibold"
                  }`}
              >
                {item.label}
              </span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
