export const AVAILABLE_ACCOUNT_ICONS = [
  "fbi",
  "hdfc",
  "sbi",
] as const;

export const ACCOUNT_COLORS = [
  "#6366F1",
  "#07B682",
  "#F59E0B",
  "#EC4899",
  "#8B5CF6",
] as const;

export function getAccountIconKey(
  iconKey: string | null,
  name: string,
) {
  const key = iconKey?.toLowerCase().trim();
  const accountName = name.toLowerCase().trim();

  return AVAILABLE_ACCOUNT_ICONS.find(
    (icon) =>
      key === icon ||
      accountName.includes(icon),
  );
}
