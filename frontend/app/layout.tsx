import { Suspense } from "react";
import { PrivacyProvider } from "@/context/PrivacyContext";
import { ToastProvider } from "@/context/ToastContext";
import type { Metadata } from "next";
import { Plus_Jakarta_Sans } from "next/font/google";
import { Navbar } from "@/components/Navbar";
import { MainLayout } from "@/components/MainLayout";
import BottomNav from "@/components/BottomNav";
import AddTransactionButton from "@/components/AddTransactionButton";
import "./globals.css";

const jakarta = Plus_Jakarta_Sans({
  variable: "--font-jakarta",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Paisa",
  description: "Personal finance tracking made simple.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html
      lang="en"
      className={`${jakarta.variable} h-full antialiased`}
    >
      <head />
      <body className="min-h-full bg-background text-foreground">
        <div className="flex min-h-full flex-col md:hidden">
          <ToastProvider>
            <PrivacyProvider>
              <Navbar />

              <MainLayout>
                {children}
              </MainLayout>
            </PrivacyProvider>
          </ToastProvider>

          <Suspense fallback={null}>
            <AddTransactionButton />
          </Suspense>
          <BottomNav />
        </div>
        <main className="fixed inset-0 z-[200] hidden items-center justify-center bg-[#FFFAF0] px-6 text-center md:flex">
          <section className="max-w-md">
            <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-amber-100 text-3xl" aria-hidden="true">
              📱
            </div>
            <h1 className="mt-6 text-2xl font-bold text-neutral-900">Paisa is designed for mobile</h1>
            <p className="mt-3 text-sm leading-6 text-neutral-600">
              This app is made for phone screens and isn’t optimized for desktop. Open Paisa on your phone or make this window narrower to continue.
            </p>
          </section>
        </main>
      </body>
    </html>
  );
}
