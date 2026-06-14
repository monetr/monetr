import type BankAccount from '@monetr/interface/models/BankAccount';
import { ID, idPrefix } from '@monetr/interface/models/ID';
import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export default class FundingSchedule {
  readonly [idPrefix] = 'fund';

  readonly fundingScheduleId: ID<FundingSchedule>;
  readonly bankAccountId: ID<BankAccount>;
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
    this.fundingScheduleId = ID.from(data.fundingScheduleId);
    this.bankAccountId = ID.from(data.bankAccountId);
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
