import { apiFetch } from "./api";

export type AllocationType = "percentage" | "fixed" | "remainder";

export type Allocation = {
  id: string;
  transaction_id: string;
  transaction_name: string;
  jar_id: string;
  jar_name: string;
  allocation_type: AllocationType;
  amount: number;
  occurred_at: string;
  created_at: string;
};

export type AllocationMonthlySummary = {
  month: string;
  total_allocated: number;
  jar_total_allocated?: number;
};

export type AllocationListResponse = {
  allocations: Allocation[];
  total: number;
  limit: number;
  offset: number;
  monthly_summary: AllocationMonthlySummary;
};

export type AllocationFilters = {
  search?: string;
  jarId?: string;
  jarName?: string;
  transactionName?: string;
  allocationType?: AllocationType;
  fromDate?: string;
  toDate?: string;
  minAmount?: number;
  maxAmount?: number;
  month?: string;
  limit?: number;
  offset?: number;
};

export async function getAllocations(
  filters: AllocationFilters = {},
): Promise<AllocationListResponse> {
  const params = new URLSearchParams();

  params.set("limit", String(filters.limit ?? 4));
  params.set("offset", String(filters.offset ?? 0));

  if (filters.search) {
    params.set("search", filters.search);
  }

  if (filters.jarId) {
    params.set("jar_id", filters.jarId);
  }

  if (filters.jarName) {
    params.set("jar_name", filters.jarName);
  }

  if (filters.transactionName) {
    params.set("transaction_name", filters.transactionName);
  }

  if (filters.allocationType) {
    params.set("allocation_type", filters.allocationType);
  }

  if (filters.fromDate) {
    params.set("from_date", filters.fromDate);
  }

  if (filters.toDate) {
    params.set("to_date", filters.toDate);
  }

  if (filters.minAmount !== undefined) {
    params.set("min_amount", String(filters.minAmount));
  }

  if (filters.maxAmount !== undefined) {
    params.set("max_amount", String(filters.maxAmount));
  }

  if (filters.month) {
    params.set("month", filters.month);
  }

  return apiFetch<AllocationListResponse>(
    `/api/v1/allocations/?${params.toString()}`,
  );
}
