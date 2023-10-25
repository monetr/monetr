import { parseJSON } from 'date-fns';


export default class FundingSchedule {
  fundingScheduleId: number;
  bankAccountId: number;
  name: string;
  description?: string;
  ruleset: string;
  lastOccurrence?: Date;
  nextOccurrence: Date;
  nextOccurrenceOriginal: Date;
  excludeWeekends: boolean;
  estimatedDeposit: number | null;

  constructor(data?: Partial<FundingSchedule>) {
    if (data) {
      Object.assign(this, {
        ...data,
        lastOccurrence: data.lastOccurrence && parseJSON(data.lastOccurrence),
        nextOccurrence: parseJSON(data.nextOccurrence),
        nextOccurrenceOriginal: parseJSON(data.nextOccurrenceOriginal),
      });
    }
  }
}
