"use client";

import { useState, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Step1Amount from "@/components/transactions/steps/Step1Amount";
import Step2Details from "@/components/transactions/steps/Step2Details";
import Step3Confirmation from "@/components/transactions/steps/Step3Confirmation";
import { TransactionType, TransactionCategory, formatRFC3339 } from "@/components/transactions/transaction-utils";

export type TransactionFormData = {
  amount: string; // Store as string for input handling, convert to number (paise) on save
  type: TransactionType;
  name: string;
  category: TransactionCategory;
  fromAccountId: string;
  fromAccountName: string;
  fromAccountBalance: number;
  fromAccountIconKey: string;
  toAccountId: string;
  toAccountName: string;
  toAccountBalance: number;
  toAccountIconKey: string;
  jarId: string;
  jarName: string;
  jarBalance: number;
  jarIconKey: string;
  occurredAt: string;
  isMasterIncome: boolean;
};

export default function AddTransactionPageWrapper() {
  return (
    <Suspense fallback={<div className="flex min-h-screen bg-[#f8f9ff]"></div>}>
      <AddTransactionPage />
    </Suspense>
  );
}

function AddTransactionPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const dateParam = searchParams.get('date');
  const [step, setStep] = useState(1);

  const getInitialOccurredAt = () => {
    const now = new Date();
    if (dateParam) {
      // A date-only value is local. Timestamp values must be converted to local
      // time first (for example, midnight in India is the prior UTC evening).
      const dateOnly = /^(\d{4})-(\d{2})-(\d{2})$/.exec(dateParam);
      const d = dateOnly
        ? new Date(Number(dateOnly[1]), Number(dateOnly[2]) - 1, Number(dateOnly[3]))
        : new Date(dateParam);
      if (!isNaN(d.getTime())) {
        d.setHours(now.getHours(), now.getMinutes(), now.getSeconds(), 0);
        return formatRFC3339(d);
      }
    }
    return formatRFC3339(now);
  };

  const [formData, setFormData] = useState<TransactionFormData>({
    amount: "0",
    type: "expense",
    name: "",
    category: "other",
    fromAccountId: "",
    fromAccountName: "",
    fromAccountBalance: 0,
    fromAccountIconKey: "",
    toAccountId: "",
    toAccountName: "",
    toAccountBalance: 0,
    toAccountIconKey: "",
    jarId: "",
    jarName: "",
    jarBalance: 0,
    jarIconKey: "",
    occurredAt: getInitialOccurredAt(),
    isMasterIncome: false,
  });

  const updateFormData = (data: Partial<TransactionFormData>) => {
    setFormData((prev) => ({ ...prev, ...data }));
  };

  const nextStep = () => setStep((prev) => prev + 1);
  const prevStep = () => setStep((prev) => Math.max(1, prev - 1));
  const handleCancel = () => {
    if (typeof window !== "undefined" && window.history.length > 1) {
      router.back();
    } else {
      router.push("/transactions");
    }
  };

  return (
    <div className="flex min-h-screen flex-col bg-[#f8f9ff]">
      {step === 1 && (
        <Step1Amount
          formData={formData}
          updateFormData={updateFormData}
          onNext={nextStep}
          onCancel={handleCancel}
        />
      )}
      {step === 2 && (
        <Step2Details
          formData={formData}
          updateFormData={updateFormData}
          onNext={nextStep}
          onBack={prevStep}
          onCancel={handleCancel}
        />
      )}
      {step === 3 && (
        <Step3Confirmation
          formData={formData}
          updateFormData={updateFormData}
          onBack={prevStep}
          onCancel={handleCancel}
          onComplete={(addAnother, occurredAt) => {
            if (addAnother) {
              setFormData({
                amount: "0",
                type: "expense",
                name: "",
                category: "other",
                fromAccountId: "",
                fromAccountName: "",
                fromAccountBalance: 0,
                fromAccountIconKey: "",
                toAccountId: "",
                toAccountName: "",
                toAccountBalance: 0,
                toAccountIconKey: "",
                jarId: "",
                jarName: "",
                jarBalance: 0,
                jarIconKey: "",
                occurredAt,
                isMasterIncome: false,
              });
              setStep(1);
            } else {
              router.push("/transactions");
            }
          }}
        />
      )}
    </div>
  );
}
