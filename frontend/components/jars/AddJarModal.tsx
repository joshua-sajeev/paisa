"use client";

import { useState } from "react";

import { CloseFill } from "@material-symbols-svg/react/icons/close";

import { apiFetch } from "@/lib/api";

type AllocationType = "percentage" | "fixed" | "remainder";

type Props = {
  open: boolean;
  onCloseAction: () => void;
  onCreatedAction: () => void;
};

type CreateJarData = {
  name: string;
  allocation_type: AllocationType;
  allocation_value: number;
  icon_key?: string;
};

export default function AddJarModal({
  open,
  onCloseAction,
  onCreatedAction,
}: Props) {
  const [name, setName] = useState("");
  const [allocationType, setAllocationType] =
    useState<AllocationType>("remainder");
  const [allocationValue, setAllocationValue] = useState("");
  const [iconKey, setIconKey] = useState("");
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

    let value = 0;

    if (allocationType !== "remainder") {
      value = Number(allocationValue);

      if (!Number.isFinite(value) || value <= 0) {
        setError(
          allocationType === "percentage"
            ? "Enter a percentage greater than 0."
            : "Enter a fixed amount greater than 0.",
        );
        return;
      }

      if (allocationType === "percentage" && value > 100) {
        setError("Percentage cannot be greater than 100.");
        return;
      }
    }

    setSubmitting(true);
    setError("");

    const data: CreateJarData = {
      name: trimmedName,
      allocation_type: allocationType,
      allocation_value: value,
    };

    if (trimmedIconKey) {
      data.icon_key = trimmedIconKey;
    }

    try {
      const result = await apiFetch<unknown>("/api/v1/jars", {
        method: "POST",
        body: JSON.stringify(data),
      });

      if (!result) {
        setError("Unable to create jar.");
        setSubmitting(false);
        return;
      }

      setName("");
      setAllocationType("remainder");
      setAllocationValue("");
      setIconKey("");
      setSubmitting(false);

      onCreatedAction();
      onCloseAction();
    } catch (error) {
      console.error("Failed to create jar:", error);

      setError("Unable to connect to the server.");
      setSubmitting(false);
    }
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
        {/* Header */}
        <div className="mb-5 flex items-center justify-between">
          <h2 className="text-xl font-bold text-on-surface">
            Add Jar
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

        {/* Jar Name */}
        <label
          htmlFor="jar-name"
          className="mb-1 block text-sm font-semibold text-on-surface"
        >
          Jar name
        </label>

        <input
          id="jar-name"
          value={name}
          onChange={(event) => setName(event.target.value)}
          placeholder="e.g. Necessities"
          className="mb-4 w-full rounded-xl bg-surface-container-lowest px-4 py-3 text-on-surface outline-none ring-1 ring-outline-variant/50 transition focus:ring-2 focus:ring-primary"
        />

        {/* Allocation Type */}
        <label
          htmlFor="allocation-type"
          className="mb-1 block text-sm font-semibold text-on-surface"
        >
          Allocation type
        </label>

        <select
          id="allocation-type"
          value={allocationType}
          onChange={(event) => {
            setAllocationType(
              event.target.value as AllocationType,
            );
            setAllocationValue("");
            setError("");
          }}
          className="mb-4 w-full appearance-none rounded-xl bg-surface-container-lowest px-4 py-3 text-on-surface outline-none ring-1 ring-outline-variant/50 transition focus:ring-2 focus:ring-primary"
        >
          <option value="remainder">Remainder</option>
          <option value="percentage">Percentage</option>
          <option value="fixed">Fixed amount</option>
        </select>

        {/* Allocation Value */}
        {allocationType !== "remainder" && (
          <>
            <label
              htmlFor="allocation-value"
              className="mb-1 block text-sm font-semibold text-on-surface"
            >
              {allocationType === "percentage"
                ? "Percentage"
                : "Fixed amount"}
            </label>

            <div className="relative mb-4">
              {allocationType === "fixed" && (
                <span className="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-sm font-semibold text-secondary">
                  ₹
                </span>
              )}

              <input
                id="allocation-value"
                type="number"
                min="0"
                max={
                  allocationType === "percentage"
                    ? 100
                    : undefined
                }
                step={
                  allocationType === "percentage"
                    ? "0.01"
                    : "1"
                }
                value={allocationValue}
                onChange={(event) =>
                  setAllocationValue(event.target.value)
                }
                placeholder={
                  allocationType === "percentage"
                    ? "e.g. 30"
                    : "e.g. 5000"
                }
                className={`w-full rounded-xl bg-surface-container-lowest py-3 pr-4 text-on-surface outline-none ring-1 ring-outline-variant/50 transition focus:ring-2 focus:ring-primary ${allocationType === "fixed"
                    ? "pl-9"
                    : "pl-4"
                  }`}
              />

              {allocationType === "percentage" && (
                <span className="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-sm font-semibold text-secondary">
                  %
                </span>
              )}
            </div>
          </>
        )}

        {/* Icon Key */}
        <label
          htmlFor="jar-icon-key"
          className="mb-1 block text-sm font-semibold text-on-surface"
        >
          Icon key
          <span className="ml-1 font-normal text-secondary">
            Optional
          </span>
        </label>

        <input
          id="jar-icon-key"
          type="text"
          value={iconKey}
          onChange={(event) => setIconKey(event.target.value)}
          placeholder="e.g. necessities"
          className="mb-5 w-full rounded-xl bg-surface-container-lowest px-4 py-3 text-on-surface outline-none ring-1 ring-outline-variant/50 transition focus:ring-2 focus:ring-primary"
        />

        {/* Error */}
        {error && (
          <p className="mb-4 text-sm font-medium text-error">
            {error}
          </p>
        )}

        {/* Submit */}
        <button
          type="button"
          onClick={handleSubmit}
          disabled={!name.trim() || submitting}
          className="w-full rounded-xl bg-[#2563eb] py-3 font-semibold text-white transition active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-50"
        >
          {submitting ? "Creating..." : "Create Jar"}
        </button>
      </div>
    </div>
  );
}
