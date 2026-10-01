import AccountCard from "./AccountCard";
import { Account } from "./AccountsPage";

type Props = {
  title: string;
  accounts: Account[];
  onUpdatedAction?: () => void;
};

export default function AccountSection({
  title,
  accounts,
  onUpdatedAction,
}: Props) {
  return (
    <section className="mb-6">
      <h2 className="mb-3 text-sm font-bold uppercase tracking-wider text-on-surface-variant">
        {title}
      </h2>

      <div className="space-y-3">
        {accounts.map((account, index) => (
          <AccountCard
            key={account.id}
            account={account}
            index={index}
            onUpdatedAction={onUpdatedAction}
          />
        ))}
      </div>
    </section>
  );
}
