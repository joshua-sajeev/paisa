"use client";
import { usePathname } from "next/navigation";

export function MainLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const noBottomNavPaths = ["/login", "/transactions/add"];
  const hasBottomNav = !noBottomNavPaths.includes(pathname) && !pathname.endsWith("/edit");

  return (
    <main className={`flex-1 ${hasBottomNav ? "pb-24" : ""}`}>
      {children}
    </main>
  );
}
