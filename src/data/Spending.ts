import { Moment } from "moment";
import { parseToMoment, parseToMomentMaybe } from "util/parseToMoment";

export enum SpendingType {
  Expense = 0,
  Goal = 1,
}

export interface SpendingFields {
  spendingId: number;
  bankAccountId: number;
  fundingScheduleId?: number;
  name: string;
  description?: string;
  spendingType: SpendingType;
  targetAmount: number;
  currentAmount: number;
  recurrenceRule: string;
  lastRecurrence?: Moment;
  nextRecurrence: Moment;
  nextContributionAmount: number;
  isBehind: boolean;
}

export default class Spending implements SpendingFields {
  spendingId: number;
  bankAccountId: number;
  fundingScheduleId?: number;
  name: string;
  description?: string;
  spendingType: SpendingType;
  targetAmount: number;
  currentAmount: number;
  recurrenceRule: string;
  lastRecurrence?: Moment;
  nextRecurrence: Moment;
  nextContributionAmount: number;
  isBehind: boolean;

  constructor(data?: SpendingFields) {
    if (data) {
      Object.assign(this, {
        ...data,
        lastRecurrence: parseToMomentMaybe(data.lastRecurrence),
        nextRecurrence: parseToMoment(data.nextRecurrence),
      });
    }
  }

  getTargetAmountString(): string {
    return `$${ (this.targetAmount / 100).toFixed(2) }`;
  }

  getCurrentAmountString(): string {
    return `$${ (this.currentAmount / 100).toFixed(2) }`;
  }

  getIsExpense(): boolean {
    return this.spendingType === SpendingType.Expense;
  }
}
