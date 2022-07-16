import formatAmount from 'util/formatAmount';

export default class Balance {
  bankAccountId: number;
  available: number;
  current: number;
  safe: number;
  expenses: number;
  goals: number;

  constructor(data?: Partial<Balance>) {
    if (data) {
      Object.assign(this, data);
    }
  }

  getSafeToSpendString(): string {
    return formatAmount(this.safe);
  }

  getAvailableString(): string {
    return formatAmount(this.available);
  }

  getExpensesString(): string {
    return formatAmount(this.expenses);
  }

  getGoalsString(): string {
    return formatAmount(this.goals);
  }
}
