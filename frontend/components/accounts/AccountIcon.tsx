import Image from "next/image";

import { AccountBalanceFill } from "@material-symbols-svg/react/icons/account-balance";

import {
  ACCOUNT_COLORS,
  getAccountIconKey,
} from "@/lib/accounts/account-icons";

type Props = {
  iconKey: string;
  name: string;
  index?: number;
  size?: "sm" | "md";
};

export default function AccountIcon({
  iconKey,
  name,
  index = 0,
  size = "md",
}: Props) {
  const matchedIcon = getAccountIconKey(
    iconKey,
    name,
  );

  const dimensions =
    size === "sm"
      ? {
        container: "h-7 w-7 p-1",
        icon: 24,
        fallback: 15,
      }
      : {
        container: "h-12 w-12 p-2",
        icon: 32,
        fallback: 22,
      };

  if (matchedIcon) {
    return (
      <div
        className={`flex shrink-0 items-center justify-center overflow-hidden rounded-xl bg-white shadow-sm ${dimensions.container}`}
      >
        <Image
          src={`/icons/accounts/${matchedIcon}.svg`}
          alt={name}
          width={dimensions.icon}
          height={dimensions.icon}
          className="h-full w-full object-contain"
        />
      </div>
    );
  }

  const bgColor =
    ACCOUNT_COLORS[index % ACCOUNT_COLORS.length];

  return (
    <div
      className={`flex shrink-0 items-center justify-center rounded-xl ${dimensions.container}`}
      style={{ backgroundColor: bgColor }}
    >
      <AccountBalanceFill
        width={dimensions.fallback}
        height={dimensions.fallback}
        color="#FFFFFF"
      />
    </div>
  );
}
