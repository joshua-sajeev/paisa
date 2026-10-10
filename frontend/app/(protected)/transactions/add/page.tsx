"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Step1Amount from "@/components/transactions/steps/Step1Amount";
import Step2Details from "@/components/transactions/steps/Step2Details";
import Step3Confirmation from "@/components/transactions/steps/Step3Confirmation";
import { TransactionType, TransactionCategory } from "@/components/transactions/transaction-utils";

export type TransactionFormData = {
  amount: string; // Store as string for input handling, convert to number (paise) on save
  type: TransactionType;
  name: string;
  category: TransactionCategory;
  fromAccountId: string;
  fromAccountName: string;
  fromAccountBalance: number;
  toAccountId: string;
  toAccountName: string;
  toAccountBalance: number;
  jarId: string;
  jarName: string;
  jarBalance: number;
  occurredAt: string;
  isMasterIncome: boolean;
};

export default function AddTransactionPage() {
  const router = useRouter();
  const [step, setStep] = useState(1);
  const [formData, setFormData] = useState<TransactionFormData>({
    amount: "0",
    type: "expense",
    name: "",
    category: "other",
    fromAccountId: "",
    fromAccountName: "",
    fromAccountBalance: 0,
    toAccountId: "",
    toAccountName: "",
    toAccountBalance: 0,
    jarId: "",
    jarName: "",
    jarBalance: 0,
    occurredAt: new Date().toISOString(),
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
          onComplete={() => router.push("/transactions")}
        />
      )}
    </div>
  );
}
