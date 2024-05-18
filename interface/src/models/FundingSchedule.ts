import { parseJSON } from 'date-fns';


export default class FundingSchedule {
  fundingScheduleId: string;
  bankAccountId: string;
  name: string;
  description?: string;
  ruleset: string;
  lastRecurrence?: Date;
  nextRecurrence: Date;
  nextRecurrenceOriginal: Date;
  excludeWeekends: boolean;
  estimatedDeposit: number | null;

  constructor(data?: Partial<FundingSchedule>) {
    if (data) {
      Object.assign(this, {
        ...data,
        lastOccurrence: data.lastRecurrence && parseJSON(data.lastRecurrence),
        nextOccurrence: parseJSON(data.nextRecurrence),
        nextOccurrenceOriginal: parseJSON(data.nextRecurrenceOriginal),
      });
    }
  }
}
