export default class Balance {
  bankAccountId: number;
  available: number;
  current: number;
  safe: number;
  expenses: number;
  goals: number;

  constructor(data?: Partial<Balance>) {
    if (data) {
      Object.assign(this, data)
    }
  }

  getSafeToSpendString(): string {
    return `$${ (this.safe / 100).toFixed(2) }`;
  }

  getAvailableString(): string {
    return `$${ (this.available / 100).toFixed(2) }`;
  }

  getExpensesString(): string {
    return `$${ (this.expenses / 100).toFixed(2) }`;
  }

  getGoalsString(): string {
    return `$${ (this.goals / 100).toFixed(2) }`;
  }
}
