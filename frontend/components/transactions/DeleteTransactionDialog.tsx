"use client";

import { apiFetch } from "@/lib/api";
import { useToast } from "@/context/ToastContext";
import DeleteConfirmationDialog from "@/components/ui/DeleteConfirmationDialog";

import { type Transaction } from "./transaction-utils";

export type DeletableTransaction = Pick<
  Transaction,
  "id" | "name" | "amount" | "occurred_at"
> & {
  account?: string;
};

type Props = {
  transaction: DeletableTransaction | null;
  onClose: () => void;
  onDeleted?: (transaction: DeletableTransaction) => void | Promise<void>;
};

export default function DeleteTransactionDialog({
  transaction,
  onClose,
  onDeleted,
}: Props) {
  const { showToast } = useToast();

  async function handleConfirm() {
    if (!transaction) return;

    await apiFetch<void>(`/api/v1/transactions/${transaction.id}`, {
      method: "DELETE",
    });

    showToast("Transaction deleted");
    void (async () => {
      try {
        await onDeleted?.(transaction);
      } catch (cause) {
        console.error("Failed to refresh after deleting transaction", cause);
      }
    })();
  }

  return (
    <DeleteConfirmationDialog
      open={transaction !== null}
      title="Delete transaction?"
      description={
        <>
          Are you sure you want to delete{" "}
          <strong className="font-bold text-slate-900">
            {transaction?.name ? `“${transaction.name}”` : "this transaction"}
          </strong>
          ?
        </>
      }
      confirmLabel="Delete"
      onConfirm={handleConfirm}
      onClose={onClose}
    />
  );
}
