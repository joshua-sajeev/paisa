"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

import { ArchiveFillW700 } from "@material-symbols-svg/react/icons/archive";
import { ChevronRight } from "@material-symbols-svg/react/icons/chevron-right";
import { CloseFill } from "@material-symbols-svg/react/icons/close";

import { updateAccount } from "@/lib/accounts/accounts";
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

  const formattedBalance = new Intl.NumberFormat(
    "en-IN",
    {
      style: "currency",
      currency: "INR",
      minimumFractionDigits: 2,
    },
  ).format(account.balance);

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
              className={`text-lg font-extrabold tabular-nums tracking-tight ${
                account.balance < 0
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
        account={account}
        open={showEditModal}
        onCloseAction={() => setShowEditModal(false)}
        onUpdatedAction={() =>
          onUpdatedAction?.()
        }
      />

      {showArchiveModal && (
        <div
          className="fixed inset-0 z-[100] flex items-center justify-center bg-black/30 px-4 backdrop-blur-sm"
          onClick={() => {
            if (!archiving) {
              setShowArchiveModal(false);
            }
          }}
        >
          <div
            className="w-full max-w-sm rounded-2xl bg-white p-5 shadow-2xl"
            onClick={(event) =>
              event.stopPropagation()
            }
          >
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-lg font-bold text-on-surface">
                Archive account?
              </h2>

              <button
                type="button"
                disabled={archiving}
                onClick={() =>
                  setShowArchiveModal(false)
                }
                className="flex h-9 w-9 items-center justify-center rounded-full transition hover:bg-surface-container disabled:opacity-50"
                aria-label="Close"
              >
                <CloseFill
                  width={20}
                  height={20}
                />
              </button>
            </div>

            <p className="text-sm leading-6 text-secondary">
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

            <div className="mt-5 flex gap-3">
              <button
                type="button"
                disabled={archiving}
                onClick={() =>
                  setShowArchiveModal(false)
                }
                className="flex-1 rounded-xl bg-surface-container py-3 text-sm font-semibold text-on-surface transition hover:bg-surface-container-high disabled:opacity-50"
              >
                Cancel
              </button>

              <button
                type="button"
                disabled={archiving}
                onClick={handleArchive}
                className="flex-1 rounded-xl bg-[#2563eb] py-3 text-sm font-semibold text-white transition active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-50"
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
