export interface DashboardSummary {
  total_balance: number;
  monthly_income: number;
  monthly_expense: number;
  monthly_savings: number;
  monthly_savings_rate: number;
}

export interface DashboardAccount {
  id: string;
  name: string;
  icon_key: string;
  balance: number;
  is_primary: boolean;
}

export interface DashboardJar {
  id: string;
  name: string;
  allocation_type: "percentage" | "fixed" | "remainder";
  allocation_value: number;
  allocated: number;
  used: number;
  available: number;
  used_percentage: number;
  available_percentage: number;
}

export interface DashboardGoal {
  id: string;
  name: string;
  target: number;
  contributed: number;
  remaining: number;
  progress: number;
  deadline: string;
}

export interface DashboardTransaction {
  id: string;
  name: string;
  date: string;
  type: "income" | "expense";
  amount: number;
  jar_name: string | null;
  account: string;
  account_balance: number;
  category: string;
}

export interface DashboardData {
  summary: DashboardSummary;
  accounts: DashboardAccount[];
  jars: DashboardJar[];
  goals: DashboardGoal[];
  recent_transactions: DashboardTransaction[];
}

function getDashboardUrl() {
  const configuredBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL;

  if (configuredBaseUrl) {
    return new URL("/api/v1/dashboard", configuredBaseUrl).toString();
  }

  return "/api/v1/dashboard";
}

export async function getDashboardData(): Promise<DashboardData | null> {
  try {
    const response = await fetch(getDashboardUrl(), {
      credentials: "include",
      cache: "no-store",
    });

    if (!response.ok) {
      return null;
    }

    const data = (await response.json()) as DashboardData;

    return data;
  } catch {
    return null;
  }
}
