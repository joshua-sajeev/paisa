export class ApiError extends Error {
  code: string;
  status: number;

  constructor(status: number, message: string, code: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

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
    let message = `API request failed: ${response.status}`;
    let code = "ERR_UNKNOWN";

    try {
      const text = await response.text();
      try {
        const body = JSON.parse(text);
        if (body && typeof body.message === "string") {
          message = body.message;
        }
        if (body && typeof body.code === "string") {
          code = body.code;
        }
      } catch {
        if (text && text.trim()) {
          message = text.trim();
        }
      }
    } catch {
      // Response body was not readable — keep the generic message.
    }

    throw new ApiError(response.status, message, code);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
}
