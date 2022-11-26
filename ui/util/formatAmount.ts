const enUS = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
});

export default function formatAmount(amount: number): string {
  const actual = +((amount / 100).toFixed(2));
  return enUS.format(actual);
}
