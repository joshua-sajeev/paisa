"use client";

import { useState } from "react";
import { apiFetch } from "@/lib/api";

type Jar = {
  id: string;
  name: string;
  icon_key: string;
  is_archived: boolean;
};

// Export the Props type so it's recognized properly when imported
export type EditJarModalProps = {
  open: boolean;
  jar: Jar | null;
  onCloseAction: () => void;
  onUpdatedAction: () => void;
};

export default function EditJarModal({
  open,
  jar,
  onCloseAction,
  onUpdatedAction,
}: EditJarModalProps) {
  const [name, setName] = useState(() => jar?.name ?? "");
  const [iconKey, setIconKey] = useState(() => jar?.icon_key ?? "");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!open || !jar) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const payload: { name?: string; icon_key?: string } = {};

      const trimmedName = name.trim();
      if (trimmedName && trimmedName !== jar.name) {
        payload.name = trimmedName;
      }

      const trimmedIconKey = iconKey.trim();
      if (trimmedIconKey && trimmedIconKey !== jar.icon_key) {
        payload.icon_key = trimmedIconKey;
      }

      if (Object.keys(payload).length === 0) {
        onCloseAction();
        return;
      }

      await apiFetch(`/api/v1/jars/${jar.id}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });

      onUpdatedAction();
      onCloseAction();
    } catch (err: unknown) {
      console.error("Failed to update jar:", err);
      const errorMessage =
        err instanceof Error ? err.message : "Failed to update jar. Please try again.";
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 px-4"
      onClick={onCloseAction}
    >
      <div
        className="w-full max-w-md rounded-2xl bg-white p-6 shadow-xl"
        onClick={(event) => event.stopPropagation()}
      >
        <h2 className="text-base font-bold text-slate-900">Edit Jar</h2>
        <p className="mt-1 text-xs text-slate-500">
          Update the details for <span className="font-semibold text-slate-700">{jar.name}</span>.
        </p>

        {error && (
          <div className="mt-4 rounded-lg bg-rose-50 p-3 text-xs text-rose-600 border border-rose-200">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="mt-4 space-y-4">
          <div>
            <label className="block text-xs font-semibold text-slate-700">
              Jar Name <span className="font-normal text-slate-400">(Optional)</span>
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              maxLength={100}
              className="mt-1.5 w-full rounded-lg border border-slate-200 px-3 py-2 text-xs text-slate-900 outline-none focus:border-slate-400 focus:ring-1 focus:ring-slate-400"
              placeholder="e.g. Vacation Savings"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-700">
              Icon Key <span className="font-normal text-slate-400">(Optional)</span>
            </label>
            <input
              type="text"
              value={iconKey}
              onChange={(e) => setIconKey(e.target.value)}
              maxLength={50}
              className="mt-1.5 w-full rounded-lg border border-slate-200 px-3 py-2 text-xs text-slate-900 outline-none focus:border-slate-400 focus:ring-1 focus:ring-slate-400"
              placeholder="e.g. jar, necessities, leisure"
            />
          </div>

          <div className="mt-6 flex justify-end gap-2">
            <button
              type="button"
              onClick={onCloseAction}
              disabled={loading}
              className="rounded-lg px-3 py-2 text-xs font-semibold text-slate-600 transition hover:bg-slate-50"
            >
              Cancel
            </button>

            <button
              type="submit"
              disabled={loading}
              className="rounded-lg bg-slate-900 px-4 py-2 text-xs font-semibold text-white transition hover:bg-slate-800 disabled:opacity-50"
            >
              {loading ? "Saving..." : "Save Changes"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
