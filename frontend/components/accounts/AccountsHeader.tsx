import { AddFill } from "@material-symbols-svg/react/icons/add";

type Props = {
  onAdd: () => void;
};

export default function AccountsHeader({
  onAdd,
}: Props) {
  return (
    <div className="flex items-center justify-between gap-4 py-4">
      <div className="min-w-0">
        <h1 className="text-2xl font-extrabold tracking-tight text-on-surface">
          Accounts
        </h1>
      </div>

      <button
        type="button"
        onClick={onAdd}
        className="inline-flex h-10 shrink-0 items-center gap-1.5 rounded-full bg-[#2563eb] px-4 text-sm font-semibold text-white shadow-[0_4px_12px_rgba(37,99,235,0.25)] transition active:scale-95"
      >
        <AddFill width={18} height={18} />
        <span>Add Account</span>
      </button>
    </div>
  );
}
