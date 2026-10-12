import { AccountBalanceWallet } from "@material-symbols-svg/react/w400";

type Props = {
  onClear: () => void;
};

export default function TransactionsEmpty({ onClear }: Props) {
  return (
    <div className="rounded-2xl bg-white px-6 py-12 text-center">
      <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-[#dbe1ff] text-[#004ac6]">
        <AccountBalanceWallet className="h-8 w-8" />
      </div>

      <h3 className="mt-4 text-lg font-bold text-[#0b1c30]">
        No matching transactions
      </h3>

      <p className="mx-auto mt-2 max-w-xs text-sm text-[#737686]">
        Try changing your search or filters.
      </p>

      <button
        type="button"
        onClick={onClear}
        className="mt-5 rounded-full bg-[#004ac6] px-5 py-2.5 text-xs font-bold text-white"
      >
        Clear filters
      </button>
    </div>
  );
}
