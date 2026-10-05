export async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const configuredBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL;

  const url = configuredBaseUrl
    ? new URL(path, configuredBaseUrl).toString()
    : path;

  const response = await fetch(url, {
    ...options,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
  });

  if (!response.ok) {
    throw new Error(`API request failed: ${response.status}`);
  }

  return response.json() as Promise<T>;
}
