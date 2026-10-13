"use client";

import { useEffect, useId, useRef, useState, type ReactNode } from "react";

export type DeleteDetail = {
  label: string;
  value: ReactNode;
};

export type DeleteConfirmationDialogProps = {
  open: boolean;
  /** Short question, e.g. "Delete transaction?" */
  title: string;
  /** One or two sentences explaining the consequences. */
  description: ReactNode;
  /** Identifies exactly what is being deleted (name, amount, date, ...). */
  details?: DeleteDetail[];
  /** Specific action label, e.g. "Delete transaction". */
  confirmLabel?: string;
  /** Shown on the confirm button while the request is in flight. */
  loadingLabel?: string;
  cancelLabel?: string;
  /**
   * Performs the deletion. Reject/throw to keep the dialog open and show the
   * error message; resolve to close it.
   */
  onConfirm: () => Promise<void> | void;
  onClose: () => void;
};

const FALLBACK_ERROR = "Something went wrong. Please try again.";

export default function DeleteConfirmationDialog(
  props: DeleteConfirmationDialogProps,
) {
  // Mount the dialog only while open so its loading/error state resets on every open.
  if (!props.open) return null;
  return <DialogContent {...props} />;
}

function DialogContent({
  title,
  description,
  details,
  confirmLabel = "Delete",
  loadingLabel = "Deleting…",
  cancelLabel = "Cancel",
  onConfirm,
  onClose,
}: DeleteConfirmationDialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const cancelRef = useRef<HTMLButtonElement>(null);
  const busyRef = useRef(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const titleId = useId();
  const descriptionId = useId();

  // Open as a true modal: focus is trapped, the page behind is inert, and
  // Escape is wired up natively. Restore focus and scrolling on close.
  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;

    const previouslyFocused = document.activeElement as HTMLElement | null;
    const previousOverflow = document.body.style.overflow;

    document.body.style.overflow = "hidden";
    if (!dialog.open) dialog.showModal();
    // Focus the safe option first so a stray Enter never deletes anything.
    cancelRef.current?.focus();

    return () => {
      document.body.style.overflow = previousOverflow;
      if (previouslyFocused?.isConnected) previouslyFocused.focus();
    };
  }, []);

  // After a failed attempt the buttons re-enable; put focus back on Cancel.
  useEffect(() => {
    if (error) cancelRef.current?.focus();
  }, [error]);

  function requestClose() {
    if (busyRef.current) return;
    onClose();
  }

  async function handleConfirm() {
    if (busyRef.current) return;
    busyRef.current = true;
    setLoading(true);
    setError(null);

    try {
      await onConfirm();
    } catch (cause) {
      setError(
        cause instanceof Error && cause.message ? cause.message : FALLBACK_ERROR,
      );
      busyRef.current = false;
      setLoading(false);
      return;
    }

    // Success: stay in the loading state until the parent unmounts us, so the
    // buttons never flash back to enabled.
    onClose();
  }

  return (
    <dialog
      ref={dialogRef}
      role="alertdialog"
      aria-modal="true"
      aria-labelledby={titleId}
      aria-describedby={descriptionId}
      aria-busy={loading}
      // Escape: always take over from the browser so React state stays the
      // source of truth, and ignore it while a request is in flight.
      onCancel={(event) => {
        event.preventDefault();
        requestClose();
      }}
      // Some browsers close the dialog on a repeated Escape despite
      // preventDefault. Re-open if we are mid-request, otherwise sync state.
      onClose={() => {
        const dialog = dialogRef.current;
        if (busyRef.current) {
          if (dialog && !dialog.open) dialog.showModal();
        } else {
          onClose();
        }
      }}
      // The dialog has no padding of its own, so a click whose target is the
      // dialog itself can only be a click on the backdrop.
      onClick={(event) => {
        if (event.target === event.currentTarget) requestClose();
      }}
      className="m-auto max-h-[calc(100dvh-2rem)] w-[calc(100%-2rem)] max-w-sm overflow-y-auto rounded-3xl border border-slate-200 bg-white p-0 text-slate-900 shadow-xl backdrop:bg-slate-900/40 animate-dialog-in"
    >
      <div className="p-6">
        <h2 id={titleId} className="text-lg font-bold text-slate-900">
          {title}
        </h2>
        <p id={descriptionId} className="mt-1.5 text-sm leading-6 text-slate-500">
          {description}
        </p>

        {details && details.length > 0 && (
          <dl className="mt-4 divide-y divide-slate-200/70 rounded-2xl border border-slate-200 bg-slate-50 px-4">
            {details.map((detail) => (
              <div
                key={detail.label}
                className="flex items-baseline justify-between gap-4 py-2.5 text-sm"
              >
                <dt className="shrink-0 text-slate-500">{detail.label}</dt>
                <dd className="min-w-0 break-words text-right font-semibold text-slate-900">
                  {detail.value}
                </dd>
              </div>
            ))}
          </dl>
        )}

        {error && (
          <p
            role="alert"
            className="mt-4 rounded-xl border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-700"
          >
            {error}
          </p>
        )}

        <div className="mt-6 flex flex-row items-center justify-end gap-3">
          <button
            ref={cancelRef}
            type="button"
            onClick={requestClose}
            disabled={loading}
            className="flex-1 h-11 rounded-full bg-slate-100 px-5 text-sm font-semibold text-slate-700 transition hover:bg-slate-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-400 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {cancelLabel}
          </button>

          <button
            type="button"
            onClick={handleConfirm}
            disabled={loading}
            className="flex-1 flex h-11 items-center justify-center gap-2 rounded-full bg-rose-600 px-5 text-sm font-bold text-white shadow-sm transition hover:bg-rose-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-rose-500 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-70"
          >
            {loading && (
              <span
                aria-hidden="true"
                className="h-4 w-4 animate-spin rounded-full border-2 border-white/40 border-t-white"
              />
            )}
            {loading ? loadingLabel : confirmLabel}
          </button>
        </div>
      </div>
    </dialog>
  );
}
