"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { PinKeypad } from "@/components/ui/pin-keypad";

export default function LoginPage() {
  const router = useRouter();

  const [pin, setPin] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleLogin() {
    if (pin.length !== 6 || loading) {
      return;
    }

    setLoading(true);
    setError("");

    try {
      const response = await fetch("/auth/login", {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          pin,
        }),
      });

      if (!response.ok) {
        setError("Invalid PIN");
        setPin("");
        return;
      }

      router.push("/dashboard");
    } catch {
      setError("Unable to connect to server");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="relative flex h-dvh w-full flex-col overflow-hidden bg-[#F8F9FE] text-[#111322]">
      {/* Ambient background */}
      <div className="pointer-events-none fixed inset-0 overflow-hidden">
        <div className="absolute -left-20 -top-24 h-80 w-80 animate-float-1 rounded-full bg-gradient-to-br from-[#FFDEE2] to-[#FFB7B2] opacity-60 blur-3xl mix-blend-multiply" />

        <div className="absolute -right-24 top-1/4 h-88 w-88 animate-float-2 rounded-full bg-gradient-to-bl from-[#E0DBFF] via-[#CFEBFF] to-[#D5FFF2] opacity-70 blur-3xl mix-blend-multiply" />

        <div className="absolute -bottom-20 left-1/2 h-96 w-96 -translate-x-1/2 rounded-full bg-gradient-to-t from-[#FFF3D6] via-[#FFE3EC] to-transparent opacity-70 blur-2xl" />

        {/* Decorative accents */}
        <div className="absolute right-10 top-16 h-3.5 w-3.5 animate-float-1 rounded-full bg-[#FFB800] opacity-60" />

        <div className="absolute left-8 top-36 h-4 w-4 animate-float-2 rotate-45 rounded-md bg-[#6C47FF] opacity-30" />

        <div className="absolute bottom-40 left-6 h-3 w-3 animate-float-1 rounded-full bg-[#05D69E] opacity-50" />

        <div className="absolute bottom-60 right-8 h-4 w-2 animate-float-2 rotate-12 rounded-full bg-[#FF5C67] opacity-40" />
      </div>

      {/* Login content */}
      <div className="relative z-10 mx-auto flex h-full w-full max-w-[420px] flex-col px-6 py-6">
        {/* Header */}
        <header className="pt-12 text-center">
          <div className="text-2xl font-black tracking-[-0.06em]">
            PAISA
          </div>

          <div className="mt-10">
            <h1 className="flex items-center justify-center gap-1.5 text-[28px] font-extrabold leading-tight tracking-tight">
              Welcome back
              <span className="inline-block -rotate-12 text-2xl animate-pulse-dot">
                👋
              </span>
            </h1>

            <p className="mt-1 text-sm font-medium text-slate-500">
              Enter your 6-digit security PIN
            </p>
          </div>
        </header>

        {/* PIN section */}
        <section className="mt-auto flex w-full flex-col items-center pb-2">
          {error && (
            <p className="mb-4 text-sm font-medium text-red-500">
              {error}
            </p>
          )}

          <PinKeypad
            pin={pin}
            setPinAction={setPin}
            maxLength={6}
            disabled={loading}
            onCompleteAction={handleLogin}
          />
        </section>
      </div>
    </main>
  );
}
