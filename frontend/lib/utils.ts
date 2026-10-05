export function formatMoney(paisa: number): string {
  const rupees = paisa / 100;
  return rupees.toLocaleString("en-IN", {
    maximumFractionDigits: 2,
  });
}

export function formatMonth(month: string) {
  const [year, monthNumber] = month.split("-");

  return new Intl.DateTimeFormat("en-IN", {
    month: "long",
    year: "numeric",
  }).format(new Date(Number(year), Number(monthNumber) - 1, 1));
}

export function formatDate(dateString: string) {
  const date = new Date(dateString);

  return {
    day: date.toLocaleDateString("en-IN", {
      day: "2-digit",
    }),
    month: date.toLocaleDateString("en-IN", {
      month: "short",
    }),
  };
}
