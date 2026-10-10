import { Suspense } from "react";
import { PrivacyProvider } from "@/context/PrivacyContext";
import type { Metadata } from "next";
import { Plus_Jakarta_Sans } from "next/font/google";
import { Navbar } from "@/components/Navbar";
import { MainLayout } from "@/components/MainLayout";
import  BottomNav  from "@/components/BottomNav";
import  AddTransactionButton  from "@/components/AddTransactionButton";
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
      <body className="min-h-full flex flex-col bg-background text-foreground">
        <PrivacyProvider>
          <Navbar />

          <MainLayout>
            {children}
          </MainLayout>
        </PrivacyProvider>

        <Suspense fallback={null}>
          <AddTransactionButton />
        </Suspense>
        <BottomNav />
      </body>
    </html>
  );
}
