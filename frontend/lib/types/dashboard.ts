export type DashboardSummary = {
  total_balance: number;
  monthly_income: number;
  monthly_expense: number;
  monthly_savings: number;
  monthly_savings_change: number;
};

export type DashboardAccount = {
  id: string;
  name: string;
  balance: number;
};

export type DashboardJar = {
  id: string;
  name: string;
  allocation_type: "percentage" | "fixed" | "remainder";
  allocation_value: number;
  allocated: number;
  used: number;
  available: number;
  used_percentage: number;
  available_percentage: number;
};

export type DashboardGoal = {
  id: string;
  name: string;
  target_amount: number;
  current_amount: number;
  progress_percentage: number;
};

export type DashboardTransaction = {
  id: string;
  name: string;
  date: string;
  type: "income" | "expense" | "transfer";
  amount: number;
  jar_name: string | null;
  account: string;
  account_balance: number | null;
  category: string;
};

export type DashboardData = {
  summary: DashboardSummary;
  accounts: DashboardAccount[];
  jars: DashboardJar[];
  goals: DashboardGoal[];
  recent_transactions: DashboardTransaction[];
};
