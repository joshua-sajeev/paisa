"use client";

import { useEffect, useState } from "react";

import AccountsHeader from "./AccountsHeader";
import TotalBalanceCard from "./TotalBalanceCard";
import AccountSection from "./AccountSection";
import ArchivedAccounts from "./ArchivedAccounts";
import AddAccountModal from "./AddAccountModal";

import {
  getAccounts,
  type Account as ApiAccount,
} from "@/lib/accounts/accounts";

export type Account = {
  id: string;
  name: string;
  balance: number;
  iconKey: string | null;
  isPrimary: boolean;
  isArchived: boolean;
  updatedAt: string;
};

function mapAccount(account: ApiAccount): Account {
  return {
    id: account.id,
    name: account.name,
    balance: account.balance,
    iconKey: account.icon_key,
    isPrimary: account.is_primary,
    isArchived: account.is_archived,
    updatedAt: account.updated_at,
  };
}

export default function AccountsPage() {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [showAddModal, setShowAddModal] = useState(false);
  const [loading, setLoading] = useState(true);

  async function loadAccounts() {
    const data = await getAccounts();

    if (data) {
      setAccounts(data.map(mapAccount));
    }
  }

  useEffect(() => {
    async function load() {
      const data = await getAccounts();

      if (data) {
        setAccounts(data.map(mapAccount));
      }

      setLoading(false);
    }

    load();
  }, []);

  const activeAccounts = accounts
    .filter((account) => !account.isArchived)
    .sort(
      (a, b) =>
        Number(b.isPrimary) - Number(a.isPrimary),
    );

  const archivedAccounts = accounts.filter(
    (account) => account.isArchived,
  );

  const totalBalance = activeAccounts.reduce(
    (total, account) => total + account.balance,
    0,
  );

  return (
    <>
      <main className="min-h-full bg-surface p-3 max-w-2xl mx-auto flex flex-col gap-3">
        <div className="flex flex-col gap-3 w-full pb-16">
          <AccountsHeader
            onAdd={() => setShowAddModal(true)}
          />

          {loading ? (
            <div className="py-10 text-center text-sm text-secondary">
              Loading accounts...
            </div>
          ) : (
            <>
              <TotalBalanceCard
                balance={totalBalance}
                activeCount={activeAccounts.length}
                archivedCount={archivedAccounts.length}
              />

              <AccountSection
                title="Active Accounts"
                accounts={activeAccounts}
                onUpdatedAction={loadAccounts}
              />

              <ArchivedAccounts
                accounts={archivedAccounts}
                onUpdatedAction={loadAccounts}
              />
            </>
          )}
        </div>
      </main>

      <AddAccountModal
        open={showAddModal}
        onCloseAction={() => setShowAddModal(false)}
        onCreatedAction={loadAccounts}
      />
    </>
  );
}
