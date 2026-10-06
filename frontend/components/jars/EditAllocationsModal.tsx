"use client";

import React, { useEffect, useMemo, useState } from "react";
import { ApiError } from "@/lib/api";
import { getJarIcon } from "@/components/jars/jar-icons";

export type AllocationType = "percentage" | "fixed" | "remainder";

export interface JarAllocation {
  id: string;
  name: string;
  subtitle: string;
  iconKey: string;
  iconBgColor: string;
  iconTextColor: string;
  badgeText: string;
  badgeColor: string;
  allocationType: AllocationType;
  allocationValue: number;
}

interface EditAllocationsModalProps {
  isOpen: boolean;
  onClose: () => void;
  initialJars: JarAllocation[];
  onSave: (updatedJars: JarAllocation[]) => Promise<void>;
}

type ValidationError =
  | "EMPTY"
  | "NO_REMAINDER"
  | "MULTIPLE_REMAINDERS"
  | "PERCENTAGE_EXCEEDS_LIMIT"
  | "INVALID_ALLOCATION_VALUE"
  | "INVALID_ALLOCATION_TYPE"
  | null;

function getServerErrorMessage(code: string, fallback: string): string {
  switch (code) {
    case "ERR_INVALID_ALLOCATION":
      // The backend message is already human-readable:
      // "must have exactly 1 remainder jar"
      // "cannot have multiple remainder jars"
      // "percentages must sum to less than 100%"
      // "at least one jar is required"
      return fallback;

    case "ERR_INVALID_ALLOCATION_TYPE":
      return "One or more jars have an invalid allocation type.";

    case "ERR_INVALID_ALLOCATION_VALUE":
      return "One or more allocation values are invalid.";

    case "ERR_JAR_NOT_FOUND":
      return "A jar was not found. It may have been deleted. Please refresh and try again.";

    case "ERR_INVALID_INPUT":
      return "The allocation request is invalid. Please check your inputs.";

    default:
      return fallback || "Failed to update allocations.";
  }
}

export const EditAllocationsModal: React.FC<
  EditAllocationsModalProps
> = ({ isOpen, onClose, initialJars, onSave }) => {
  const [jars, setJars] = useState<JarAllocation[]>(initialJars);
  const [loading, setLoading] = useState(false);
  const [toastMessage, setToastMessage] = useState<string | null>(null);
  const [serverError, setServerError] = useState<string | null>(null);

  useEffect(() => {
    if (!isOpen) return;

    setJars(initialJars);
    setServerError(null);
    setToastMessage(null);
  }, [initialJars, isOpen]);

  const stats = useMemo(() => {
    let totalPercent = 0;
    let remainderCount = 0;

    jars.forEach((jar) => {
      if (jar.allocationType === "percentage") {
        totalPercent += Number(jar.allocationValue) || 0;
      }

      if (jar.allocationType === "remainder") {
        remainderCount += 1;
      }
    });

    return {
      totalPercent,
      remainderCount,
    };
  }, [jars]);

  const validationError = useMemo<ValidationError>(() => {
    if (jars.length === 0) {
      return "EMPTY";
    }

    for (const jar of jars) {
      if (
        jar.allocationType !== "percentage" &&
        jar.allocationType !== "fixed" &&
        jar.allocationType !== "remainder"
      ) {
        return "INVALID_ALLOCATION_TYPE";
      }

      /*
       * Remainder does not use an allocation value.
       */
      if (jar.allocationType === "remainder") {
        continue;
      }

      const value = Number(jar.allocationValue);

      if (!Number.isFinite(value) || value < 0) {
        return "INVALID_ALLOCATION_VALUE";
      }

      /*
       * Percentage is represented as a normal number:
       * 25 = 25%
       */
      if (jar.allocationType === "percentage" && value > 100) {
        return "PERCENTAGE_EXCEEDS_LIMIT";
      }

      /*
       * Fixed allocation is stored in paise.
       * Negative values are never valid.
       */
      if (jar.allocationType === "fixed" && value < 0) {
        return "INVALID_ALLOCATION_VALUE";
      }
    }

    if (stats.remainderCount === 0) {
      return "NO_REMAINDER";
    }

    if (stats.remainderCount > 1) {
      return "MULTIPLE_REMAINDERS";
    }

    if (stats.totalPercent > 100) {
      return "PERCENTAGE_EXCEEDS_LIMIT";
    }

    return null;
  }, [jars, stats]);

  const validationMessage = useMemo(() => {
    switch (validationError) {
      case "EMPTY":
        return {
          title: "No allocation configuration",
          message: "At least one active jar is required.",
        };

      case "NO_REMAINDER":
        return {
          title: "Remainder jar required",
          message:
            "Exactly one active jar must be configured as Remainder.",
        };

      case "MULTIPLE_REMAINDERS":
        return {
          title: "Multiple Remainder jars",
          message:
            "Only one active jar can be configured as Remainder.",
        };

      case "PERCENTAGE_EXCEEDS_LIMIT":
        return {
          title: "Percentage allocation exceeds 100%",
          message:
            "The combined percentage allocation cannot exceed 100%.",
        };

      case "INVALID_ALLOCATION_VALUE":
        return {
          title: "Invalid allocation value",
          message:
            "Allocation values must be zero or greater.",
        };

      case "INVALID_ALLOCATION_TYPE":
        return {
          title: "Invalid allocation type",
          message:
            "One or more jars have an unsupported allocation type.",
        };

      default:
        return null;
    }
  }, [validationError]);

  const sortedJars = useMemo(() => {
    return [...jars].sort((a, b) => {
      if (
        a.allocationType === "remainder" &&
        b.allocationType !== "remainder"
      ) {
        return -1;
      }

      if (
        a.allocationType !== "remainder" &&
        b.allocationType === "remainder"
      ) {
        return 1;
      }

      return 0;
    });
  }, [jars]);

  if (!isOpen) return null;

  const handleTypeChange = (
    id: string,
    newType: AllocationType,
  ) => {
    setServerError(null);

    setJars((prev) =>
      prev.map((jar) => {
        if (jar.id !== id) return jar;

        let defaultVal = jar.allocationValue;

        /*
         * A value that was valid as fixed/percentage may not make
         * sense when switching types.
         */
        if (newType === "percentage") {
          /*
           * Fixed values are stored in paise, so don't carry a
           * potentially huge paise value into percentage mode.
           */
          if (jar.allocationType === "fixed") {
            defaultVal = 10;
          }

          if (defaultVal > 100) {
            defaultVal = 10;
          }
        }

        if (newType === "fixed") {
          /*
           * Percentage values are not paise.
           * Start fixed allocations from zero when switching.
           */
          if (jar.allocationType === "percentage") {
            defaultVal = 0;
          }
        }

        if (newType === "remainder") {
          defaultVal = 0;
        }

        return {
          ...jar,
          allocationType: newType,
          allocationValue: defaultVal,
        };
      }),
    );
  };

  const handleValueChange = (
    id: string,
    val: number,
  ) => {
    setServerError(null);

    setJars((prev) =>
      prev.map((jar) =>
        jar.id === id
          ? {
            ...jar,
            allocationValue: val,
          }
          : jar,
      ),
    );
  };

  const handleFixedAmountChange = (
    id: string,
    rupees: string,
  ) => {
    setServerError(null);

    if (rupees === "") {
      handleValueChange(id, 0);
      return;
    }

    const parsed = Number(rupees);

    if (!Number.isFinite(parsed) || parsed < 0) {
      return;
    }

    handleValueChange(id, Math.round(parsed * 100));
  };

  const handleSaveClick = async () => {
    setServerError(null);

    /*
     * Never submit a configuration that we already know is invalid.
     */
    if (validationError) {
      return;
    }

    setLoading(true);

    try {
      await onSave(jars);

      setToastMessage("Allocations updated successfully!");

      setTimeout(() => {
        setToastMessage(null);
        onClose();
      }, 1200);
    } catch (error) {
      if (error instanceof ApiError) {
        setServerError(getServerErrorMessage(error.code, error.message));
      } else {
        setServerError("Failed to update allocations.");
      }
    } finally {
      setLoading(false);
    }
  };

  const hasError = Boolean(validationError || serverError);

  return (
    <div className="fixed inset-x-0 bottom-[80px] top-0 z-50 flex flex-col justify-end bg-slate-900/60 p-0 backdrop-blur-sm animate-in fade-in duration-200 sm:bottom-0 sm:justify-center sm:p-3">
      <section
        aria-labelledby="modalTitle"
        aria-modal="true"
        role="dialog"
        className="flex max-h-[92%] w-full flex-col overflow-hidden rounded-t-[32px] border border-slate-100 bg-white shadow-2xl sm:rounded-3xl"
      >
        <div
          aria-hidden="true"
          className="flex w-full justify-center pb-1 pt-3 sm:hidden"
        >
          <div className="h-1.5 w-12 rounded-full bg-slate-200" />
        </div>

        <header className="flex-shrink-0 border-b border-slate-100 px-5 pb-3 pt-3">
          <div className="flex items-start justify-between">
            <div>
              <div className="flex items-center gap-2">
                <h2
                  className="text-lg font-bold tracking-tight text-slate-900"
                  id="modalTitle"
                >
                  Edit Allocation Rules
                </h2>

                <span className="inline-flex items-center rounded-full border border-blue-200 bg-blue-50 px-2 py-0.5 text-[10px] font-semibold text-blue-700">
                  Auto-Split
                </span>
              </div>

              <p className="mt-0.5 text-xs text-slate-500">
                Configure monthly auto-distribution rules across
                active jars
              </p>
            </div>

            <button
              aria-label="Close modal"
              onClick={onClose}
              type="button"
              className="-mr-1 -mt-1 flex h-8 w-8 items-center justify-center rounded-full text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600"
            >
              <svg
                className="h-5 w-5"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                viewBox="0 0 24 24"
              >
                <path
                  d="M6 18L18 6M6 6l12 12"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            </button>
          </div>
        </header>

        <main className="flex-1 space-y-3 overflow-y-auto px-5 py-3">
          {sortedJars.length === 0 ? (
            <div className="py-8 text-center text-xs text-slate-400">
              No active jars found.
            </div>
          ) : (
            sortedJars.map((jar) => {
              const IconComponent = getJarIcon(jar.iconKey);

              return (
                <article
                  key={jar.id}
                  className="rounded-2xl border border-slate-200/90 bg-white p-3.5 shadow-sm transition-all"
                >
                  <div className="mb-3 flex items-center justify-between">
                    <div className="flex items-center gap-2.5">
                      <div
                        className={`flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-xl border ${jar.badgeColor}`}
                      >
                        <IconComponent size={20} />
                      </div>

                      <div>
                        <h3 className="text-sm font-bold leading-tight text-slate-800">
                          {jar.name}
                        </h3>

                        <p className="text-[11px] text-slate-400">
                          {jar.subtitle}
                        </p>
                      </div>
                    </div>

                    <span
                      className={`rounded-full border px-2 py-0.5 text-[11px] font-semibold ${jar.badgeColor}`}
                    >
                      {jar.badgeText}
                    </span>
                  </div>

                  <div className="grid grid-cols-12 gap-2 border-t border-slate-100 pt-2">
                    <div className="col-span-6">
                      <label className="mb-1 block text-[10px] font-bold uppercase tracking-wider text-slate-500">
                        Rule Type
                      </label>

                      <select
                        value={jar.allocationType}
                        onChange={(e) =>
                          handleTypeChange(
                            jar.id,
                            e.target.value as AllocationType,
                          )
                        }
                        className="w-full rounded-lg border border-slate-200 bg-slate-50 px-2.5 py-2 text-xs font-semibold text-slate-700 outline-none transition focus:border-transparent focus:ring-2 focus:ring-blue-500"
                      >
                        <option value="remainder">
                          Remainder
                        </option>

                        <option value="percentage">
                          Percentage (%)
                        </option>

                        <option value="fixed">
                          Fixed Amount (₹)
                        </option>
                      </select>
                    </div>

                    <div className="col-span-6">
                      <label className="mb-1 block text-[10px] font-bold uppercase tracking-wider text-slate-500">
                        Allocation Value
                      </label>

                      {jar.allocationType === "remainder" ? (
                        <div className="flex h-[34px] items-center rounded-lg border border-emerald-200/80 bg-emerald-50/70 px-2.5 text-[11px] font-semibold text-emerald-700">
                          <span className="truncate">
                            Unallocated balance
                          </span>
                        </div>
                      ) : (
                        <div className="relative rounded-lg">
                          {jar.allocationType === "fixed" && (
                            <span className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-2.5 text-xs font-semibold text-slate-400">
                              ₹
                            </span>
                          )}

                          <input
                            type="number"
                            min="0"
                            step={
                              jar.allocationType === "fixed"
                                ? "0.01"
                                : "1"
                            }
                            max={
                              jar.allocationType === "percentage"
                                ? 100
                                : undefined
                            }
                            value={
                              jar.allocationType === "fixed"
                                ? jar.allocationValue / 100
                                : jar.allocationValue
                            }
                            onChange={(e) => {
                              if (
                                jar.allocationType ===
                                "fixed"
                              ) {
                                handleFixedAmountChange(
                                  jar.id,
                                  e.target.value,
                                );
                                return;
                              }

                              const value = Number(
                                e.target.value,
                              );

                              handleValueChange(
                                jar.id,
                                Number.isFinite(value)
                                  ? value
                                  : 0,
                              );
                            }}
                            className={`w-full rounded-lg border border-slate-200 bg-white py-2 text-xs font-semibold text-slate-800 outline-none focus:border-transparent focus:ring-2 focus:ring-blue-500 ${jar.allocationType === "fixed"
                                ? "pl-6 pr-2.5"
                                : "px-2.5 pr-7"
                              }`}
                            placeholder="0"
                          />

                          {jar.allocationType ===
                            "percentage" && (
                              <span className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2.5 text-xs font-semibold text-slate-400">
                                %
                              </span>
                            )}
                        </div>
                      )}
                    </div>
                  </div>
                </article>
              );
            })
          )}

          {validationMessage && (
            <div className="flex items-start gap-2 rounded-xl border border-rose-200 bg-rose-50 p-3 text-xs text-rose-700">
              <svg
                className="mt-0.5 h-4 w-4 flex-shrink-0"
                fill="currentColor"
                viewBox="0 0 20 20"
              >
                <path
                  fillRule="evenodd"
                  d="M18 10A8 8 0 112 10a8 8 0 0116 0zM9 8a1 1 0 112 0v4a1 1 0 11-2 0V8zm1-5a1 1 0 100 2 1 1 0 000-2z"
                  clipRule="evenodd"
                />
              </svg>

              <div>
                <strong className="block">
                  {validationMessage.title}
                </strong>

                <p className="mt-0.5 text-[11px] leading-relaxed">
                  {validationMessage.message}
                </p>
              </div>
            </div>
          )}

          {serverError && !validationMessage && (
            <div className="flex items-start gap-2 rounded-xl border border-rose-200 bg-rose-50 p-3 text-xs text-rose-700">
              <svg
                className="mt-0.5 h-4 w-4 flex-shrink-0"
                fill="currentColor"
                viewBox="0 0 20 20"
              >
                <path
                  fillRule="evenodd"
                  d="M18 10A8 8 0 112 10a8 8 0 0116 0zM9 8a1 1 0 112 0v4a1 1 0 11-2 0V8zm1-5a1 1 0 100 2 1 1 0 000-2z"
                  clipRule="evenodd"
                />
              </svg>

              <div>
                <strong className="block">
                  Unable to save allocations
                </strong>

                <p className="mt-0.5 text-[11px] leading-relaxed">
                  {serverError}
                </p>
              </div>
            </div>
          )}

          {!hasError && (
            <div className="flex items-start gap-2 rounded-xl border border-slate-200 bg-slate-100/80 p-3 text-xs text-slate-600">
              <svg
                className="mt-0.5 h-4 w-4 flex-shrink-0 text-blue-600"
                fill="currentColor"
                viewBox="0 0 20 20"
              >
                <path
                  fillRule="evenodd"
                  d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                  clipRule="evenodd"
                />
              </svg>

              <p className="text-[11px] leading-relaxed">
                <strong>Rule Logic:</strong> Exactly 1 Remainder
                jar collects remaining funds. Percentages cannot
                exceed 100%. Fixed cuts are deducted first.
              </p>
            </div>
          )}
        </main>

        <footer className="flex flex-shrink-0 items-center gap-3 border-t border-slate-100 bg-white p-4">
          <button
            onClick={onClose}
            type="button"
            disabled={loading}
            className="flex-1 rounded-xl bg-slate-100 px-4 py-3 text-center text-xs font-bold text-slate-700 transition-all hover:bg-slate-200 active:scale-95 disabled:cursor-not-allowed disabled:opacity-50"
          >
            Cancel
          </button>

          <button
            disabled={Boolean(validationError) || loading}
            onClick={handleSaveClick}
            type="button"
            className={`flex flex-[2] items-center justify-center gap-2 rounded-xl bg-blue-600 px-4 py-3 text-xs font-bold text-white shadow-md shadow-blue-500/20 transition-all hover:bg-blue-700 active:scale-95 ${validationError || loading
                ? "cursor-not-allowed opacity-50"
                : ""
              }`}
          >
            <span>
              {loading ? "Saving..." : "Save Allocations"}
            </span>

            {loading && (
              <svg
                className="h-4 w-4 animate-spin text-white"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle
                  className="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  strokeWidth="4"
                />

                <path
                  className="opacity-75"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  fill="currentColor"
                />
              </svg>
            )}
          </button>
        </footer>
      </section>

      {toastMessage && (
        <div className="fixed left-1/2 top-12 z-50 flex -translate-x-1/2 items-center gap-2 rounded-full bg-slate-900 px-4 py-2.5 text-xs text-white shadow-lg animate-in slide-in-from-top-4 fade-in duration-300">
          <span className="h-2 w-2 rounded-full bg-emerald-400" />
          <span>{toastMessage}</span>
        </div>
      )}
    </div>
  );
};
