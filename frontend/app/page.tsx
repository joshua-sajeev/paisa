"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

export default function Home() {
  const router = useRouter();

  useEffect(() => {
    async function checkSession() {
      try {
        const response = await fetch("/api/v1/session", {
          credentials: "include",
          cache: "no-store",
        });

        if (response.ok) {
          router.replace("/dashboard");
        } else {
          router.replace("/login");
        }
      } catch {
        router.replace("/login");
      }
    }

    checkSession();
  }, [router]);

  return null;
}
