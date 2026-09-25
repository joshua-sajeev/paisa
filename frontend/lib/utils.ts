export function formatMoney(paisa: number): string {
  const rupees = paisa / 100;
  return rupees.toLocaleString('en-IN', {
    maximumFractionDigits: 2,
  });
}
