import parseDate from '@monetr/interface/util/parseDate';

export default class FundingSchedule {
  fundingScheduleId: string;
  bankAccountId: string;
  name: string;
  description?: string;
  ruleset: string;
  lastRecurrence?: Date;
  nextRecurrence: Date;
  readonly nextRecurrenceOriginal: Date;
  excludeWeekends: boolean;
  estimatedDeposit: number | null;

  constructor(data?: Partial<FundingSchedule>) {
    if (data) {
      Object.assign(this, {
        ...data,
        lastRecurrence: parseDate(data?.lastRecurrence),
        nextRecurrence: parseDate(data?.nextRecurrence),
        nextRecurrenceOriginal: parseDate(data?.nextRecurrenceOriginal),
      });
    }
  }
}
