import { format, isThisYear } from 'date-fns';

import type BankAccount from '@monetr/interface/models/BankAccount';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import { ID } from '@monetr/interface/models/ID';
import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export enum SpendingType {
  FreeToUse = -1, // Cannot be present on actual responses!
  Expense = 0,
  Goal = 1,
}

export default class Spending {
  readonly spendingId: ID<Spending>;
  readonly bankAccountId: ID<BankAccount>;
  fundingScheduleId: ID<FundingSchedule>;
  name: string;
  description: string | null;
  readonly spendingType: SpendingType;
  targetAmount: number;
  currentAmount: number;
  readonly usedAmount: number;
  ruleset: string | null;
  readonly lastRecurrence: Date | null;
  nextRecurrence: Date;
  nextContributionAmount: number;
  readonly isBehind: boolean;
  isPaused: boolean;
  autoCreateTransaction: boolean;
  readonly createdAt: Date;

  constructor(data: WithJsonValues<Spending>) {
    this.spendingId = ID.from<Spending>(data.spendingId);
    this.bankAccountId = ID.from<BankAccount>(data.bankAccountId);
    this.fundingScheduleId = ID.from<FundingSchedule>(data.fundingScheduleId);
    this.name = data.name;
    this.description = data.description;
    this.spendingType = data.spendingType;
    this.targetAmount = data.targetAmount;
    this.currentAmount = data.currentAmount;
    this.usedAmount = data.usedAmount;
    this.ruleset = data.ruleset;
    this.lastRecurrence = data.lastRecurrence ? parseDate(data.lastRecurrence) : null;
    this.nextRecurrence = parseDate(data.nextRecurrence);
    this.nextContributionAmount = data.nextContributionAmount;
    this.isBehind = data.isBehind;
    this.isPaused = data.isPaused;
    this.autoCreateTransaction = data.autoCreateTransaction;
    this.createdAt = parseDate(data.createdAt);
  }

  // getNextOccurrence string will return a friendly date string representing the next time this spending object is due.
  // If the next time the spending object is due is a different year than the current one; then the year will be
  // appended to the end of the date string.
  getNextOccurrenceString(): string {
    if (!this.nextRecurrence) {
      return '';
    }

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
