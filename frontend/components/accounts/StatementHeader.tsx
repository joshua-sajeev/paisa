"use client";

import { ArrowDownwardFillW700 } from "@material-symbols-svg/react/icons/arrow-downward";
import { ArrowUpwardFillW700 } from "@material-symbols-svg/react/icons/arrow-upward";
import { DownloadFillW700 } from "@material-symbols-svg/react/icons/download";
import { StarFillW700 } from "@material-symbols-svg/react/icons/star";

import type { StatementAccount } from "./AccountStatement";

type Props = {
  account: StatementAccount;
  inflow: number;
  outflow: number;
  net: number;
};

function formatMoney(paise: number) {
  return new Intl.NumberFormat("en-IN", {
    style: "currency",
    currency: "INR",
    maximumFractionDigits: 0,
  }).format(paise / 100);
}

export default function StatementHeader({
  account,
  inflow,
  outflow,
  net,
}: Props) {
  return (
    <section className="mb-5">
      <div className="mb-4 flex items-center justify-between">
        <a
          href="/accounts"
          className="text-sm font-semibold text-black/60"
        >
          ← Back to Accounts
        </a>

        <button
          type="button"
          className="flex h-10 w-10 items-center justify-center rounded-xl border border-black/10 bg-white"
          aria-label="Download statement"
        >
          <DownloadFillW700 className="h-[21px] w-[21px]" />
        </button>
      </div>

      <h1 className="mb-4 text-2xl font-bold tracking-tight">
        {account.name} Statement
      </h1>

      <div className="overflow-hidden rounded-3xl border border-black/10 bg-white">
        <div className="p-5">
          <div className="flex items-start justify-between">
            <div className="mt-1">
              <p className="text-xs font-medium text-black/45">
                Available Balance
              </p>

              <p className="mt-1 text-3xl font-black tracking-tight">
                {formatMoney(account.balance)}
              </p>
            </div>

            <div className="flex flex-wrap justify-end gap-1.5">
              {account.is_primary && (
                <span className="flex items-center gap-1 rounded-full bg-[#f7cd1b]/20 px-2.5 py-1 text-[11px] font-bold text-black">
                  <StarFillW700 className="h-3 w-3" />
                  Primary
                </span>
              )}

              <span
                className={`rounded-full px-2.5 py-1 text-[11px] font-bold ${account.is_archived
                  ? "bg-black/5 text-black/45"
                  : "bg-[#07b682]/10 text-[#07845f]"
                  }`}
              >
                {account.is_archived ? "Archived" : "Active"}
              </span>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-3 border-t border-black/10">
          <div className="p-4">
            <div className="mb-1 flex items-center gap-1 text-xs font-medium text-black/45">
              <ArrowDownwardFillW700 className="h-4 w-4" />
              Inflow
            </div>

            <p className="text-sm font-bold text-[#07845f]">
              {formatMoney(inflow)}
            </p>
          </div>

          <div className="border-x border-black/10 p-4">
            <div className="mb-1 flex items-center gap-1 text-xs font-medium text-black/45">
              <ArrowUpwardFillW700 className="h-4 w-4" />
              Outflow
            </div>

            <p className="text-sm font-bold text-[#d95757]">
              {formatMoney(outflow)}
            </p>
          </div>

          <div className="p-4">
            <p className="mb-1 text-xs font-medium text-black/45">
              Net Gain
            </p>

            <p
              className={`text-sm font-bold ${net >= 0 ? "text-[#07845f]" : "text-[#d95757]"
                }`}
            >
              {net >= 0 ? "+" : ""}
              {formatMoney(net)}
            </p>
          </div>
        </div>
      </div>
    </section>
  );
}
