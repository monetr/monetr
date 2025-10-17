import { format, isThisYear } from 'date-fns';

import parseDate from '@monetr/interface/util/parseDate';

export enum SpendingType {
  Expense = 0,
  Goal = 1,
}

export default class Spending {
  spendingId: string;
  bankAccountId: string;
  fundingScheduleId: string;
  name: string;
  description?: string;
  spendingType: SpendingType;
  targetAmount: number;
  currentAmount: number;
  usedAmount: number;
  ruleset: string | null;
  lastRecurrence: Date | null;
  nextRecurrence: Date | null;
  nextContributionAmount: number;
  isBehind: boolean;
  isPaused: boolean;
  dateCreated: Date | null;

  constructor(data?: Partial<Spending>) {
    if (data) {
      Object.assign(this, {
        ...data,
        lastRecurrence: parseDate(data.lastRecurrence),
        nextRecurrence: parseDate(data.nextRecurrence),
        dateCreated: parseDate(data.dateCreated),
      });
    }
  }

  // getNextOccurrence string will return a friendly date string representing the next time this spending object is due.
  // If the next time the spending object is due is a different year than the current one; then the year will be
  // appended to the end of the date string.
  getNextOccurrenceString(): string {
    return isThisYear(this.nextRecurrence)
      ? format(this.nextRecurrence, 'MMM do')
      : format(this.nextRecurrence, 'MMM do, yyyy');
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

  getGoalSavedAmount(): number {
    return this.currentAmount + this.usedAmount;
  }
}
