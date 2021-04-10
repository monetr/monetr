import { Moment } from "moment";
import { parseToMoment, parseToMomentMaybe } from "util/parseToMoment";

export enum SpendingType {
  Expense = 0,
  Goal = 1,
}

export default class Spending {
  spendingId: number;
  bankAccountId: number;
  fundingScheduleId?: number;
  name: string;
  description?: string;
  spendingType: SpendingType;
  targetAmount: number;
  currentAmount: number;
  usedAmount: number;
  recurrenceRule: string;
  lastRecurrence?: Moment;
  nextRecurrence: Moment;
  nextContributionAmount: number;
  isBehind: boolean;

  constructor(data?: Partial<Spending>) {
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

  getNextContributionAmountString(): string {
    return `$${ (this.nextContributionAmount / 100).toFixed(2) }`;
  }

  getIsExpense(): boolean {
    return this.spendingType === SpendingType.Expense;
  }

  getIsGoal(): boolean {
    return this.spendingType === SpendingType.Goal;
  }

  // getGoalIsInProgress will return true if the user is still contributing to the goal. This is determined by looking
  // at what is currently allocated to the goal plus what has already been used on the goal. If the sum of these two
  // values is less than the target amount for the goal then we are still contributing to the goal.
  getGoalIsInProgress(): boolean {
    return this.currentAmount + this.usedAmount < this.targetAmount;
  }

  getGoalSavedAmountString(): string {
    return `$${ ((this.currentAmount + this.usedAmount) / 100).toFixed(2) }`;
  }
}
