
export default function formatAmount(amount: number): string {
  return `$${(amount / 100).toFixed(2)}`;
}
