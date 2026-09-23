export async function apiFetch(
  path: string,
  options: RequestInit = {},
) {
  const configuredBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL;
  const url = configuredBaseUrl
    ? new URL(path, configuredBaseUrl).toString()
    : path;

  return fetch(url, {
    ...options,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
  });
}
