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
        lastRecurrence: data.lastRecurrence && parseJSON(data.lastRecurrence),
        nextRecurrence: parseJSON(data.nextRecurrence),
        nextRecurrenceOriginal: parseJSON(data.nextRecurrenceOriginal),
      });
    }
  }
}
