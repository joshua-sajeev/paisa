type SessionResponse = {
  authenticated?: boolean;
};

function getSessionUrl() {
  const configuredBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL;

  if (configuredBaseUrl) {
    return new URL("/api/v1/session", configuredBaseUrl).toString();
  }

  return "/api/v1/session";
}

export async function getAuthenticatedSession() {
  try {
    const response = await fetch(getSessionUrl(), {
      credentials: "include",
      cache: "no-store",
    });

    if (!response.ok) {
      return false;
    }

    const session = (await response.json()) as SessionResponse;

    return session.authenticated === true;
  } catch {
    return false;
  }
}
