import { formatAmount } from '@monetr/interface/util/amounts';

export default class Balance {
  bankAccountId: number;
  available: number;
  current: number;
  free: number;
  expenses: number;
  goals: number;

  constructor(data?: Partial<Balance>) {
    if (data) Object.assign(this, data);
  }

  getFreeToUseString(): string {
    return formatAmount(this.free);
  }

  getAvailableString(): string {
    return formatAmount(this.available);
  }

  getCurrentString(): string {
    return formatAmount(this.current);
  }

  getExpensesString(): string {
    return formatAmount(this.expenses);
  }

  getGoalsString(): string {
    return formatAmount(this.goals);
  }
}
