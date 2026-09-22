"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { getAuthenticatedSession } from "@/lib/session";

export default function Home() {
  const router = useRouter();

  useEffect(() => {
    async function checkSession() {
      const authenticated = await getAuthenticatedSession();

      router.replace(authenticated ? "/dashboard" : "/login");
    }

    checkSession();
  }, [router]);

  return null;
}
