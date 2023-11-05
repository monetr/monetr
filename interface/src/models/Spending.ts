
import { format, isThisYear, parseJSON } from 'date-fns';

import { amountToFriendly, formatAmount } from '@monetr/interface/util/amounts';

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
        lastRecurrence: data.lastRecurrence && parseJSON(data.lastRecurrence),
        nextRecurrence: data.nextRecurrence && parseJSON(data.nextRecurrence),
        dateCreated: data.dateCreated && parseJSON(data.dateCreated),
      });
    }
  }

  // getNextOccurrence string will return a friendly date string representing the next time this spending object is due.
  // If the next time the spending object is due is a different year than the current one; then the year will be
  // appended to the end of the date string.
  getNextOccurrenceString(): string {
    return isThisYear(this.nextRecurrence) ?
      format(this.nextRecurrence, 'MMM do') :
      format(this.nextRecurrence, 'MMM do, yyyy');
  }

  getTargetAmountString(): string {
    return formatAmount(this.targetAmount);
  }

  getTargetAmountDollars(): number {
    return amountToFriendly(this.targetAmount);
  }

  getCurrentAmountString(): string {
    return formatAmount(this.currentAmount);
  }

  getUsedAmountString(): string {
    return formatAmount(this.usedAmount);
  }

  getNextContributionAmountString(): string {
    return formatAmount(this.nextContributionAmount);
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
    return (this.currentAmount + this.usedAmount) < this.targetAmount;
  }

  getGoalSavedAmountString(): string {
    return formatAmount(this.currentAmount + this.usedAmount);
  }
}
