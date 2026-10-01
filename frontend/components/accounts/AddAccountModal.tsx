"use client";

import { useState } from "react";

import { CloseFill } from "@material-symbols-svg/react/icons/close";

import { createAccount } from "@/lib/accounts/accounts";

type Props = {
  open: boolean;
  onCloseAction: () => void;
  onCreatedAction: () => void;
};

export default function AddAccountModal({
  open,
  onCloseAction,
  onCreatedAction,
}: Props) {
  const [name, setName] = useState("");
  const [iconKey, setIconKey] = useState("");
  const [isPrimary, setIsPrimary] = useState(false);
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
      name: string;
      icon_key?: string;
      is_primary?: boolean;
    } = {
      name: trimmedName,
    };

    if (trimmedIconKey) {
      data.icon_key = trimmedIconKey;
    }

    if (isPrimary) {
      data.is_primary = true;
    }

const result = await createAccount(data);

if (!result) {
  setError("Unable to connect to the server.");
  setSubmitting(false);
  return;
}

if ("error" in result) {
  setError(result.error);
  setSubmitting(false);
  return;
}

    setName("");
    setIconKey("");
    setIsPrimary(false);
    setSubmitting(false);

    onCreatedAction();
    onCloseAction();
  }

  return (
    <div
      className="fixed inset-0 z-[100] flex items-end justify-center bg-black/30 backdrop-blur-sm"
      onClick={onCloseAction}
    >
      <div
        className="w-full max-w-2xl rounded-t-3xl bg-white p-5 shadow-2xl"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="mb-5 flex items-center justify-between">
          <h2 className="text-xl font-bold text-on-surface">
            Add Account
          </h2>

          <button
            type="button"
            onClick={onCloseAction}
            className="flex h-10 w-10 items-center justify-center rounded-full transition hover:bg-surface-container"
            aria-label="Close"
          >
            <CloseFill width={20} height={20} />
          </button>
        </div>

        <label
          htmlFor="account-name"
          className="mb-1 block text-sm font-semibold text-on-surface"
        >
          Account name
        </label>

        <input
          id="account-name"
          value={name}
          onChange={(event) => setName(event.target.value)}
          placeholder="e.g. HDFC"
          className="mb-4 w-full rounded-xl bg-surface-container-lowest px-4 py-3 text-on-surface outline-none ring-1 ring-outline-variant/50 transition focus:ring-2 focus:ring-primary"
        />

        <label
          htmlFor="account-icon-key"
          className="mb-1 block text-sm font-semibold text-on-surface"
        >
          Icon key
          <span className="ml-1 font-normal text-secondary">
            Optional
          </span>
        </label>

        <input
          id="account-icon-key"
          type="text"
          value={iconKey}
          onChange={(event) => setIconKey(event.target.value)}
          placeholder="e.g. hdfc"
          className="mb-5 w-full rounded-xl bg-surface-container-lowest px-4 py-3 text-on-surface outline-none ring-1 ring-outline-variant/50 transition focus:ring-2 focus:ring-primary"
        />

        <label className="mb-5 flex cursor-pointer items-center gap-3">
          <input
            type="checkbox"
            checked={isPrimary}
            onChange={(event) =>
              setIsPrimary(event.target.checked)
            }
            className="h-4 w-4 accent-[#2563eb]"
          />

          <span className="text-sm font-semibold text-on-surface">
            Set as primary account
          </span>
        </label>

        {error && (
          <p className="mb-4 text-sm font-medium text-error">
            {error}
          </p>
        )}

        <button
          type="button"
          onClick={handleSubmit}
          disabled={!name.trim() || submitting}
          className="w-full rounded-xl bg-[#2563eb] py-3 font-semibold text-white transition active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-50"
        >
          {submitting ? "Creating..." : "Create Account"}
        </button>
      </div>
    </div>
  );
}
