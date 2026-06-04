import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export default class FundingSchedule {
  readonly fundingScheduleId: string;
  readonly bankAccountId: string;
  name: string;
  description: string | null;
  ruleset: string;
  readonly lastRecurrence: Date | null;
  nextRecurrence: Date;
  readonly nextRecurrenceOriginal: Date;
  excludeWeekends: boolean;
  autoCreateTransaction: boolean;
  estimatedDeposit: number | null;

  constructor(data: WithJsonValues<FundingSchedule>) {
    this.fundingScheduleId = data.fundingScheduleId;
    this.bankAccountId = data.bankAccountId;
    this.name = data.name;
    this.description = data.description;
    this.ruleset = data.ruleset;
    this.lastRecurrence = data.lastRecurrence ? parseDate(data.lastRecurrence) : null;
    this.nextRecurrence = parseDate(data.nextRecurrence);
    this.nextRecurrenceOriginal = parseDate(data.nextRecurrenceOriginal);
    this.excludeWeekends = data.excludeWeekends;
    this.autoCreateTransaction = data.autoCreateTransaction;
    this.estimatedDeposit = data.estimatedDeposit;
  }
}
