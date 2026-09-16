import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "PAISA",
  description: "Personal finance tracker",
};

export default function RootLayout({
  children,
}: LayoutProps<"/">) {
  return (
    <html lang="en" className="h-full antialiased">
      <body className="min-h-full flex flex-col bg-[#f6f8fd]">
        {children}
      </body>
    </html>
  );
}
