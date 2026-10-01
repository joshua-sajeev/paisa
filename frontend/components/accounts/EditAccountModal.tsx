"use client";

import { useState } from "react";

import { CloseFill } from "@material-symbols-svg/react/icons/close";

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
      className="fixed inset-0 z-[110] flex items-end justify-center bg-black/30 backdrop-blur-sm"
      onClick={() => {
        if (!submitting) {
          onCloseAction();
        }
      }}
    >
      <div
        className="w-full max-w-2xl rounded-t-3xl bg-white p-5 shadow-2xl"
        onClick={(event) =>
          event.stopPropagation()
        }
      >
        <div className="mb-5 flex items-center justify-between">
          <h2 className="text-xl font-bold text-on-surface">
            Edit Account
          </h2>

          <button
            type="button"
            disabled={submitting}
            onClick={onCloseAction}
            className="flex h-10 w-10 items-center justify-center rounded-full transition hover:bg-surface-container disabled:opacity-50"
            aria-label="Close"
          >
            <CloseFill
              width={20}
              height={20}
            />
          </button>
        </div>

        <label
          htmlFor="edit-account-name"
          className="mb-1 block text-sm font-semibold text-on-surface"
        >
          Account name
        </label>

        <input
          id="edit-account-name"
          value={name}
          onChange={(event) =>
            setName(event.target.value)
          }
          disabled={submitting}
          className="mb-4 w-full rounded-xl bg-surface-container-lowest px-4 py-3 text-on-surface outline-none ring-1 ring-outline-variant/50 transition focus:ring-2 focus:ring-primary disabled:opacity-50"
        />

        <label
          htmlFor="edit-account-icon-key"
          className="mb-1 block text-sm font-semibold text-on-surface"
        >
          Icon key
          <span className="ml-1 font-normal text-secondary">
            Optional
          </span>
        </label>

        <input
          id="edit-account-icon-key"
          type="text"
          value={iconKey}
          onChange={(event) =>
            setIconKey(event.target.value)
          }
          disabled={submitting}
          placeholder="e.g. hdfc"
          className="mb-5 w-full rounded-xl bg-surface-container-lowest px-4 py-3 text-on-surface outline-none ring-1 ring-outline-variant/50 transition focus:ring-2 focus:ring-primary disabled:opacity-50"
        />

        <div className="space-y-3">
          <label className="flex cursor-pointer items-center gap-3">
            <input
              type="checkbox"
              checked={isPrimary}
              onChange={(event) =>
                setIsPrimary(event.target.checked)
              }
              disabled={submitting}
              className="h-4 w-4 accent-[#2563eb]"
            />

            <span className="text-sm font-semibold text-on-surface">
              Primary account
            </span>
          </label>

          <label className="flex cursor-pointer items-center gap-3">
            <input
              type="checkbox"
              checked={isArchived}
              onChange={(event) =>
                setIsArchived(event.target.checked)
              }
              disabled={submitting}
              className="h-4 w-4 accent-[#2563eb]"
            />

            <span className="text-sm font-semibold text-on-surface">
              Archived
            </span>
          </label>
        </div>

        {error && (
          <p className="mt-4 text-sm font-medium text-error">
            {error}
          </p>
        )}

        <button
          type="button"
          onClick={handleSubmit}
          disabled={!name.trim() || submitting}
          className="mt-5 w-full rounded-xl bg-[#2563eb] py-3 font-semibold text-white transition active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-50"
        >
          {submitting ? "Saving..." : "Save Changes"}
        </button>
      </div>
    </div>
  );
}
