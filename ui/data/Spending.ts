import moment, { Moment } from 'moment';
import { parseToMoment, parseToMomentMaybe } from 'util/parseToMoment';

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
  isPaused: boolean;
  dateCreated: Moment;

  constructor(data?: Partial<Spending>) {
    if (data) {
      Object.assign(this, {
        ...data,
        lastRecurrence: parseToMomentMaybe(data.lastRecurrence),
        nextRecurrence: parseToMoment(data.nextRecurrence),
        dateCreated: parseToMoment(data.dateCreated),
      });
    }
  }

  // getNextOccurrence string will return a friendly date string representing the next time this spending object is due.
  // If the next time the spending object is due is a different year than the current one; then the year will be
  // appended to the end of the date string.
  getNextOccurrenceString(): string {
    return this.nextRecurrence.year() === moment().year() ?
      this.nextRecurrence.format('MMM Do') :
      this.nextRecurrence.format('MMM Do, YYYY');
  }

  getTargetAmountString(): string {
    return `$${ (this.targetAmount / 100).toFixed(2) }`;
  }

  getTargetAmountDollars(): number {
    return this.targetAmount / 100;
  }

  getCurrentAmountString(): string {
    return `$${ (this.currentAmount / 100).toFixed(2) }`;
  }

  getUsedAmountString(): string {
    return `$${ (this.usedAmount / 100).toFixed(2) }`;
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
    return (this.currentAmount + this.usedAmount) < this.targetAmount;
  }

  getGoalSavedAmountString(): string {
    return `$${ ((this.currentAmount + this.usedAmount) / 100).toFixed(2) }`;
  }
}
