import { AmountType, formatAmount } from '@monetr/interface/util/amounts';

export default class Balance {
  bankAccountId: string;
  currency: string;
  available: number;
  current: number;
  free: number;
  expenses: number;
  goals: number;

  constructor(data?: Partial<Balance>) {
    if (data) Object.assign(this, data);
  }

  getFreeToUseString(locale: string = 'en_US'): string {
    return formatAmount(this.free, AmountType.Stored, locale, this.currency);
  }

  getAvailableString(locale: string = 'en_US'): string {
    return formatAmount(this.available, AmountType.Stored, locale, this.currency);
  }

  getCurrentString(locale: string = 'en_US'): string {
    return formatAmount(this.current, AmountType.Stored, locale, this.currency);
  }

  getExpensesString(locale: string = 'en_US'): string {
    return formatAmount(this.expenses, AmountType.Stored, locale, this.currency);
  }

  getGoalsString(locale: string = 'en_US'): string {
    return formatAmount(this.goals, AmountType.Stored, locale, this.currency);
  }
}
