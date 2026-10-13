"use client";

import { useState } from "react";

import { ChevronRight } from "@material-symbols-svg/react/icons/chevron-right";
import { ExpandCircleDown } from "@material-symbols-svg/react/icons/expand-circle-down";

import { updateAccount } from "@/lib/accounts/accounts";
import { formatMoney } from "@/lib/utils";

import { Account } from "./AccountsPage";
import AccountIcon from "./AccountIcon";
import EditAccountModal from "./EditAccountModal";

type Props = {
  accounts: Account[];
  onUpdatedAction?: () => void;
};

export default function ArchivedAccounts({
  accounts,
  onUpdatedAction,
}: Props) {
  const [open, setOpen] = useState(false);

  return (
    <section>
      <button
        type="button"
        onClick={() => setOpen((value) => !value)}
        className="mb-3 flex w-full items-center justify-between gap-2 px-1 text-left"
      >
        <div className="flex min-w-0 items-center gap-1.5 whitespace-nowrap">
          <ExpandCircleDown
            width={17}
            height={17}
            className={`shrink-0 text-secondary transition-transform ${
open ? "" : "-rotate-90"
}`}
          />

          <h2 className="text-base font-bold text-secondary">
            Archived Accounts
          </h2>

          <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-surface-container text-[10px] font-bold text-secondary">
            {accounts.length}
          </span>
        </div>

        <span className="shrink-0 whitespace-nowrap text-[9px] font-medium text-secondary">
          Hidden from balance
        </span>
      </button>

      {open && (
        <div className="flex flex-col gap-3">
          {accounts.map((account, index) => (
            <ArchivedAccountCard
              key={account.id}
              account={account}
              index={index}
              onUpdatedAction={onUpdatedAction}
            />
          ))}
        </div>
      )}
    </section>
  );
}

function ArchivedAccountCard({
  account,
  index,
  onUpdatedAction,
}: {
    account: Account;
    index: number;
    onUpdatedAction?: () => void;
  }) {
  const [showUnarchiveModal, setShowUnarchiveModal] =
  useState(false);
  const [unarchiving, setUnarchiving] = useState(false);
  const [error, setError] = useState("");
  const [showEditModal, setShowEditModal] = useState(false);

  const formattedBalance = `₹${formatMoney(account.balance)}`;

  function openUnarchiveModal() {
    setError("");
    setShowUnarchiveModal(true);
  }

  function closeUnarchiveModal() {
    if (!unarchiving) {
      setShowUnarchiveModal(false);
    }
  }

  async function handleUnarchive() {
    if (unarchiving) {
      return;
    }

    setUnarchiving(true);
    setError("");

    const updatedAccount = await updateAccount(account.id, {
      is_archived: false,
    });

    if (!updatedAccount) {
      setError("Failed to unarchive account.");
      setUnarchiving(false);
      return;
    }

    setShowUnarchiveModal(false);

    await onUpdatedAction?.();

    setUnarchiving(false);
  }

  return (
    <>
      <article className="rounded-2xl bg-white p-4 shadow-sm">
        <div className="flex items-center justify-between gap-3">
          <div className="flex min-w-0 items-center gap-3">
            <AccountIcon
              iconKey={account.iconKey ?? ""}
              name={account.name}
              index={index}
            />

            <span className="truncate text-lg font-bold text-on-surface">
              {account.name}
            </span>
          </div>

          <span
            className={`shrink-0 text-lg font-extrabold tabular-nums tracking-tight ${
account.balance < 0
? "text-error"
: "text-on-surface"
}`}
          >
            {formattedBalance}
          </span>
        </div>

        {error && (
          <p className="mt-2 text-xs font-medium text-error">
            {error}
          </p>
        )}

        <div className="mt-3 flex items-center justify-between">
          <div className="flex items-center gap-1">
            <button
              type="button"
              onClick={() => setShowEditModal(true)}
              className="rounded-lg px-2 py-1 text-xs font-semibold text-outline transition hover:bg-surface-container hover:text-on-surface"
            >
              Edit
            </button>
            <button
              type="button"
              onClick={openUnarchiveModal}
              disabled={unarchiving}
              className="rounded-lg px-2 py-1 text-xs font-semibold text-outline transition hover:bg-surface-container hover:text-on-surface disabled:cursor-not-allowed disabled:opacity-50"
            >
              Unarchive
            </button>
          </div>

          <button
            type="button"
            className="flex items-center gap-0.5 rounded-lg px-2 py-1 text-xs font-semibold text-[#2563eb] transition hover:bg-surface-container"
          >
            <span>Statement</span>
            <ChevronRight
              width={16}
              height={16}
            />
          </button>
        </div>
      </article>

      <EditAccountModal
        key={account.id}
        account={account}
        open={showEditModal}
        onCloseAction={() => setShowEditModal(false)}
        onUpdatedAction={() => onUpdatedAction?.()}
      />

      {showUnarchiveModal && (
        <div
          className="fixed inset-0 z-[100] flex items-center justify-center bg-black/30 px-4"
          onClick={closeUnarchiveModal}
        >
          <div
            className="w-full max-w-sm rounded-2xl bg-white p-5 shadow-xl"
            onClick={(event) =>
              event.stopPropagation()
            }
          >
            <h2 className="text-sm font-bold text-slate-900">
              Unarchive account?
            </h2>

            <p className="mt-1.5 text-xs leading-5 text-slate-500">
              Are you sure you want to unarchive{" "}
              <span className="font-semibold text-on-surface">
                {account.name}
              </span>
              ? It will become an active account and
              will be included in your total balance.
            </p>

            {error && (
              <p className="mt-3 text-sm font-medium text-error">
                {error}
              </p>
            )}

            <div className="mt-5 flex justify-end gap-2">
              <button
                type="button"
                disabled={unarchiving}
                onClick={closeUnarchiveModal}
                className="rounded-lg px-3 py-2 text-xs font-semibold text-slate-600 transition hover:bg-slate-50 disabled:opacity-50"
              >
                Cancel
              </button>

              <button
                type="button"
                disabled={unarchiving}
                onClick={handleUnarchive}
                className="rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-50"
              >
                {unarchiving
                  ? "Unarchiving..."
                  : "Unarchive"}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
