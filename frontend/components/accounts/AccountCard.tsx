"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

import { ArchiveFillW700 } from "@material-symbols-svg/react/icons/archive";
import { ChevronRight } from "@material-symbols-svg/react/icons/chevron-right";

import { updateAccount } from "@/lib/accounts/accounts";
import { formatMoney } from "@/lib/utils";
import EditAccountModal from "./EditAccountModal";
import { Account } from "./AccountsPage";
import AccountIcon from "./AccountIcon";

type Props = {
  account: Account;
  index: number;
  onUpdatedAction?: () => void;
};

export default function AccountCard({
  account,
  index,
  onUpdatedAction,
}: Props) {
  const router = useRouter();

  const [showArchiveModal, setShowArchiveModal] =
    useState(false);
  const [archiving, setArchiving] = useState(false);
  const [error, setError] = useState("");
  const [showEditModal, setShowEditModal] =
    useState(false);

  const formattedBalance = `₹${formatMoney(account.balance)}`;

  async function handleArchive() {
    if (archiving) {
      return;
    }

    setArchiving(true);
    setError("");

    const updatedAccount = await updateAccount(
      account.id,
      {
        is_archived: true,
        ...(account.isPrimary && {
          is_primary: false,
        }),
      },
    );

    if (!updatedAccount) {
      setError("Failed to archive account.");
      setArchiving(false);
      return;
    }

    setShowArchiveModal(false);

    await onUpdatedAction?.();

    setArchiving(false);
  }

  return (
    <>
      <article className="group rounded-2xl bg-white p-4 shadow-sm transition-all hover:shadow-md">
        <div className="flex items-center justify-between gap-3">
          <div className="flex min-w-0 items-center gap-3">
            <AccountIcon
              iconKey={account.iconKey ?? ""}
              name={account.name}
              index={index}
            />

            <div className="min-w-0">
              <span className="block truncate text-lg font-bold text-on-surface">
                {account.name}
              </span>

              {account.isPrimary && (
                <span className="mt-0.5 inline-block rounded-full bg-[#2563eb] px-1.5 py-0.5 text-[8px] font-bold uppercase tracking-wide text-white">
                  Primary
                </span>
              )}
            </div>
          </div>

          <div className="shrink-0 text-right">
            <div
              className={`text-lg font-extrabold tabular-nums tracking-tight ${account.balance < 0
                  ? "text-error"
                  : "text-on-surface"
                }`}
            >
              {formattedBalance}
            </div>
          </div>
        </div>

        <div className="mt-3 flex items-center justify-between">
          <div className="flex items-center gap-1">
            <button
              type="button"
              onClick={() => setShowEditModal(true)}
              className="flex items-center gap-1 rounded-lg px-2 py-1 text-xs font-semibold text-outline transition hover:bg-surface-container hover:text-on-surface"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                aria-hidden="true"
              >
                <path d="M12 20h9" />
                <path
                  d="M16.5 3.5a2.12 2.12 0 1 1 3 3L7 19l-4 1 1-4 12.5-12.5z"
                />
              </svg>

              <span>Edit</span>
            </button>

            <button
              type="button"
              onClick={() => {
                setError("");
                setShowArchiveModal(true);
              }}
              className="flex items-center gap-1 rounded-lg px-2 py-1 text-xs font-semibold text-outline transition hover:bg-surface-container hover:text-on-surface"
            >
              <ArchiveFillW700
                width={16}
                height={16}
              />

              <span>Archive</span>
            </button>
          </div>

          <button
            type="button"
            onClick={() =>
              router.push(
                `/accounts/statement?accountId=${account.id}`,
              )
            }
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
        onUpdatedAction={() =>
          onUpdatedAction?.()
        }
      />

      {showArchiveModal && (
        <div
          className="fixed inset-0 z-[100] flex items-center justify-center bg-black/30 px-4"
          onClick={() => {
            if (!archiving) {
              setShowArchiveModal(false);
            }
          }}
        >
          <div
            className="w-full max-w-sm rounded-2xl bg-white p-5 shadow-xl"
            onClick={(event) =>
              event.stopPropagation()
            }
          >
            <h2 className="text-sm font-bold text-slate-900">
              Archive account?
            </h2>

            <p className="mt-1.5 text-xs leading-5 text-slate-500">
              Are you sure you want to archive{" "}
              <span className="font-semibold text-on-surface">
                {account.name}
              </span>
              ? It will be hidden from your active
              accounts and total balance.
            </p>

            {error && (
              <p className="mt-3 text-sm font-medium text-error">
                {error}
              </p>
            )}

            <div className="mt-5 flex justify-end gap-2">
              <button
                type="button"
                disabled={archiving}
                onClick={() =>
                  setShowArchiveModal(false)
                }
                className="rounded-lg px-3 py-2 text-xs font-semibold text-slate-600 transition hover:bg-slate-50 disabled:opacity-50"
              >
                Cancel
              </button>

              <button
                type="button"
                disabled={archiving}
                onClick={handleArchive}
                className="rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-50"
              >
                {archiving
                  ? "Archiving..."
                  : "Archive"}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
