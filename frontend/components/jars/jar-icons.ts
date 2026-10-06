import {
  Movie,
  Savings,
  ReceiptLong,
  TrendingUp,
  VolunteerActivism,
} from "@material-symbols-svg/react/w400";

export const JAR_ICONS = {
  jar: Savings,
  necessities: ReceiptLong,
  leisure: Movie,
  giving: VolunteerActivism,
  investment: TrendingUp,
} as const;

export function getJarIcon(iconKey: string) {
  return JAR_ICONS[iconKey as keyof typeof JAR_ICONS] ?? Savings;
}
