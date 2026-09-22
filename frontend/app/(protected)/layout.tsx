"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getAuthenticatedSession } from "@/lib/session";

export default function ProtectedLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const router = useRouter();
  const [authenticated, setAuthenticated] = useState(false);

  useEffect(() => {
    async function checkSession() {
      const hasSession = await getAuthenticatedSession();

      if (!hasSession) {
        router.replace("/login");
        return;
      }

      setAuthenticated(true);
    }

    checkSession();
  }, [router]);

  if (!authenticated) {
    return null;
  }

  return children;
}
