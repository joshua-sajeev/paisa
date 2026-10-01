export interface Account {
  id: string;
  name: string;
  icon_key: string | null;
  balance: number;
  is_primary: boolean;
  is_archived: boolean;
  created_at: string;
  updated_at: string;
}

function getAccountsUrl() {
  const configuredBaseUrl =
    process.env.NEXT_PUBLIC_API_BASE_URL;

  if (configuredBaseUrl) {
    return new URL(
      "/api/v1/accounts",
      configuredBaseUrl,
    ).toString();
  }

  return "/api/v1/accounts";
}

export async function getAccounts(): Promise<
  Account[] | null
> {
  try {
    const response = await fetch(getAccountsUrl(), {
      credentials: "include",
      cache: "no-store",
    });

    if (!response.ok) {
      return null;
    }

    return (await response.json()) as Account[];
  } catch {
    return null;
  }
}

export async function createAccount(data: {
  name: string;
  icon_key?: string;
  is_primary?: boolean;
}): Promise<
  | { account: Account }
  | { error: string }
  | null
> {
  try {
    const response = await fetch(getAccountsUrl(), {
      method: "POST",
      headers: {
        "content-type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify(data),
    });

    if (response.ok) {
      return {
        account: (await response.json()) as Account,
      };
    }

    let message = "Failed to create account.";

    try {
      const body = (await response.json()) as {
        message?: string;
      };

      if (body.message) {
        message = body.message;
      }
    } catch {
    }

    return {
      error: message,
    };
  } catch {
    return null;
  }
}

export async function updateAccount(
  id: string,
  data: {
    name?: string;
    icon_key?: string;
    is_primary?: boolean;
    is_archived?: boolean;
  },
): Promise<Account | null> {
  try {
    const response = await fetch(
      `${getAccountsUrl()}/${id}`,
      {
        method: "PATCH",
        headers: {
          "content-type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify(data),
      },
    );

    if (!response.ok) {
      return null;
    }

    return (await response.json()) as Account;
  } catch {
    return null;
  }
}

export async function getAccountTransactions(
  accountId: string,
): Promise<unknown[] | null> {
  try {
    const response = await fetch(
      `${getAccountsUrl()}/${accountId}/transactions/`,
      {
        credentials: "include",
        cache: "no-store",
      },
    );

    if (!response.ok) {
      return null;
    }

    return (await response.json()) as unknown[];
  } catch {
    return null;
  }
}
