"use client";

import { useState } from "react";

import { updateAccount } from "@/lib/accounts/accounts";

import { Account } from "./AccountsPage";

type Props = {
  account: Account;
  open: boolean;
  onCloseAction: () => void;
  onUpdatedAction: () => void;
};

export default function EditAccountModal({
  account,
  open,
  onCloseAction,
  onUpdatedAction,
}: Props) {
  const [name, setName] = useState(account.name);
  const [iconKey, setIconKey] = useState(
    account.iconKey ?? "",
  );
  const [isPrimary, setIsPrimary] = useState(
    account.isPrimary,
  );
  const [isArchived, setIsArchived] = useState(
    account.isArchived,
  );
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  if (!open) {
    return null;
  }

  async function handleSubmit() {
    const trimmedName = name.trim();
    const trimmedIconKey = iconKey.trim();

    if (!trimmedName || submitting) {
      return;
    }

    setSubmitting(true);
    setError("");

    const data: {
      name?: string;
      icon_key?: string;
      is_primary?: boolean;
      is_archived?: boolean;
    } = {
      name: trimmedName,
      is_primary: isPrimary,
      is_archived: isArchived,
    };

    if (trimmedIconKey) {
      data.icon_key = trimmedIconKey;
    }

    const updatedAccount = await updateAccount(
      account.id,
      data,
    );

    if (!updatedAccount) {
      setError("Failed to update account.");
      setSubmitting(false);
      return;
    }

    setSubmitting(false);
    onUpdatedAction();
    onCloseAction();
  }

  return (
    <div
      className="fixed inset-0 z-[110] flex items-center justify-center bg-black/30 px-4"
      onClick={() => {
        if (!submitting) {
          onCloseAction();
        }
      }}
    >
      <div
        className="w-full max-w-md rounded-2xl bg-white p-6 shadow-xl"
        onClick={(event) => event.stopPropagation()}
      >
        <h2 className="text-lg font-bold text-slate-900">Edit Account</h2>
        <p className="mt-1 text-sm text-slate-500">
          Update the details for <span className="font-semibold text-slate-700">{account.name}</span>.
        </p>

        {error && (
          <div className="mt-4 rounded-lg border border-rose-200 bg-rose-50 p-3 text-sm text-rose-600">
            {error}
          </div>
        )}

        <div className="mt-4 space-y-4">
          <div>
            <label htmlFor="edit-account-name" className="block text-sm font-semibold text-slate-700">
              Account Name
            </label>
            <input
              id="edit-account-name"
              type="text"
              value={name}
              onChange={(event) => setName(event.target.value)}
              disabled={submitting}
              className="mt-1.5 w-full rounded-lg border border-slate-200 px-3 py-2 text-sm text-slate-900 outline-none focus:border-slate-400 focus:ring-1 focus:ring-slate-400 disabled:opacity-50"
              placeholder="e.g. Savings Account"
            />
          </div>

          <div>
            <label htmlFor="edit-account-icon-key" className="block text-sm font-semibold text-slate-700">
              Icon Key <span className="font-normal text-slate-400">(Optional)</span>
            </label>
            <input
              id="edit-account-icon-key"
              type="text"
              value={iconKey}
              onChange={(event) => setIconKey(event.target.value)}
              disabled={submitting}
              placeholder="e.g. hdfc"
              className="mt-1.5 w-full rounded-lg border border-slate-200 px-3 py-2 text-sm text-slate-900 outline-none focus:border-slate-400 focus:ring-1 focus:ring-slate-400 disabled:opacity-50"
            />
          </div>

          <div className="space-y-3 border-t border-slate-100 pt-4">
            <label className="flex cursor-pointer items-center gap-2 text-sm font-semibold text-slate-700">
              <input type="checkbox" checked={isPrimary} onChange={(event) => setIsPrimary(event.target.checked)} disabled={submitting} className="h-4 w-4 accent-slate-900" />
              Primary account
            </label>
            <label className="flex cursor-pointer items-center gap-2 text-sm font-semibold text-slate-700">
              <input type="checkbox" checked={isArchived} onChange={(event) => setIsArchived(event.target.checked)} disabled={submitting} className="h-4 w-4 accent-slate-900" />
              Archived
            </label>
          </div>
        </div>

        <div className="mt-6 flex justify-end gap-2">
          <button
            type="button"
            onClick={onCloseAction}
            disabled={submitting}
            className="rounded-lg px-3 py-2 text-sm font-semibold text-slate-600 transition hover:bg-slate-50 disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={handleSubmit}
            disabled={!name.trim() || submitting}
            className="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {submitting ? "Saving..." : "Save Changes"}
          </button>
        </div>
      </div>
    </div>
  );
}
