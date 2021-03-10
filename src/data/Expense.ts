import { Moment } from "moment";

export interface ExpenseFields {
  expenseId: number;
  bankAccountId: number;
  fundingScheduleId?: number;
  name: string;
  description?: string;
  targetAmount: number;
  currentAmount: number;
  recurrenceRule: string;
  lastRecurrence?: Moment;
  nextRecurrence: Moment;
  nextContributionAmount: number;
  isBehind: boolean;
}

export default class Expense implements ExpenseFields {
  expenseId: number;
  bankAccountId: number;
  fundingScheduleId?: number;
  name: string;
  description?: string;
  targetAmount: number;
  currentAmount: number;
  recurrenceRule: string;
  lastRecurrence?: Moment;
  nextRecurrence: Moment;
  nextContributionAmount: number;
  isBehind: boolean;

  constructor(data?: ExpenseFields) {
    if (data) {
      Object.assign(this, data);
    }
  }

  getTargetAmountString(): string {
    return `$${ (this.targetAmount / 100).toFixed(2) }`;
  }

  getCurrentAmountString(): string {
    return `$${ (this.currentAmount / 100).toFixed(2) }`;
  }
}
