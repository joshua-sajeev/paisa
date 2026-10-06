import {
  AccountBalance,
  AccountBalanceWallet,
  ArrowDownward,
  ArrowUpward,
  SwapHoriz,
} from "@material-symbols-svg/react/w400";
import { DirectionsCarFillW700 } from "@material-symbols-svg/react/icons/directions-car";
import { HomeFillW700 } from "@material-symbols-svg/react/icons/home";
import { LocalPharmacyFillW700 } from "@material-symbols-svg/react/icons/local-pharmacy";
import { MovieFillW700 } from "@material-symbols-svg/react/icons/movie";
import { ReceiptLongFillW700 } from "@material-symbols-svg/react/icons/receipt-long";
import { RestaurantFillW700 } from "@material-symbols-svg/react/icons/restaurant";
import { ShoppingCartFillW700 } from "@material-symbols-svg/react/icons/shopping-cart";
import { SwapHorizFillW700 } from "@material-symbols-svg/react/icons/swap-horiz";
import { TrendingUpFillW700 } from "@material-symbols-svg/react/icons/trending-up";
import { VolunteerActivismFillW700 } from "@material-symbols-svg/react/icons/volunteer-activism";

import { toDateKey } from "@/components/ui/DateRangePicker";

export type TransactionType = "income" | "expense" | "transfer";

export type TransactionCategory =
  | "food"
  | "transport"
  | "entertainment"
  | "groceries"
  | "health"
  | "transfer"
  | "donation"
  | "investment"
  | "housing"
  | "other";

export type Transaction = {
  id: string;
  name: string;
  occurred_at: string;
  type: TransactionType;
  amount: number;
  jar_name: string | null;
  account: string;
  account_balance: number;
  category: TransactionCategory | null;
};

export type Account = {
  id: string;
  name: string;
  is_archived?: boolean;
};

export type Jar = {
  id: string;
  name: string;
  is_archived?: boolean;
};

export type TransactionsResponse = {
  transactions: Transaction[];
  total: number;
};

export type AccountsResponse = {
  accounts: Account[];
};

export type JarsResponse = {
  jars: Jar[];
};

export const PAGE_SIZE_OPTIONS = [5, 10, 25, 50];

export const CATEGORIES: {
  value: TransactionCategory | "all";
  label: string;
}[] = [
    { value: "all", label: "All Categories" },
    { value: "food", label: "Food" },
    { value: "transport", label: "Transport" },
    { value: "entertainment", label: "Entertainment" },
    { value: "groceries", label: "Groceries" },
    { value: "health", label: "Health" },
    { value: "transfer", label: "Transfer" },
    { value: "donation", label: "Donation" },
    { value: "investment", label: "Investment" },
    { value: "housing", label: "Housing" },
    { value: "other", label: "Other" },
  ];

export function formatMoney(paise: number) {
  return new Intl.NumberFormat("en-IN", {
    style: "currency",
    currency: "INR",
    minimumFractionDigits: 2,
  }).format(paise / 100);
}

export function formatTime(value: string) {
  return new Intl.DateTimeFormat("en-IN", {
    hour: "numeric",
    minute: "2-digit",
  }).format(new Date(value));
}

export function formatDate(value: string) {
  return new Intl.DateTimeFormat("en-IN", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  }).format(new Date(value));
}

export function getDateGroup(value: string) {
  const today = new Date();

  const todayKey = toDateKey(today.toISOString());
  const dateKey = toDateKey(value);

  if (dateKey === todayKey) {
    return "Today";
  }

  const yesterday = new Date(today);
  yesterday.setDate(today.getDate() - 1);

  if (dateKey === toDateKey(yesterday.toISOString())) {
    return "Yesterday";
  }

  return formatDate(value);
}

export function getTypeClasses(type: TransactionType) {
  switch (type) {
    case "income":
      return {
        icon: "bg-[#dcfce7] text-[#006222]",
        amount: "text-[#006222]",
      };

    case "expense":
      return {
        icon: "bg-[#fee2e2] text-[#ba1a1a]",
        amount: "text-[#ba1a1a]",
      };

    case "transfer":
      return {
        icon: "bg-[#dbe1ff] text-[#004ac6]",
        amount: "text-[#004ac6]",
      };
  }
}

const CATEGORY_ICONS = {
  food: RestaurantFillW700,
  transport: DirectionsCarFillW700,
  entertainment: MovieFillW700,
  groceries: ShoppingCartFillW700,
  health: LocalPharmacyFillW700,
  transfer: SwapHorizFillW700,
  donation: VolunteerActivismFillW700,
  investment: TrendingUpFillW700,
  housing: HomeFillW700,
  other: ReceiptLongFillW700,
};

export function TransactionIcon({
  type,
  category,
}: {
  type: TransactionType;
  category?: TransactionCategory | null;
}) {
  if (category && category in CATEGORY_ICONS) {
    const Icon = CATEGORY_ICONS[category];
    return <Icon className="h-5 w-5" />;
  }

  if (type === "income") {
    return <ArrowDownward className="h-5 w-5" />;
  }

  if (type === "expense") {
    return <ArrowUpward className="h-5 w-5" />;
  }

  return <SwapHoriz className="h-5 w-5" />;
}

export function AccountIcon({ account }: { account: string }) {
  if (account.toLowerCase() === "wallet") {
    return <AccountBalanceWallet className="h-3.5 w-3.5" />;
  }

  return <AccountBalance className="h-3.5 w-3.5" />;
}

export function groupTransactions(transactions: Transaction[]) {
  const groups = new Map<string, Transaction[]>();

  for (const transaction of transactions) {
    const group = getDateGroup(transaction.occurred_at);
    const existing = groups.get(group);

    if (existing) {
      existing.push(transaction);
    } else {
      groups.set(group, [transaction]);
    }
  }

  return Array.from(groups.entries());
}
