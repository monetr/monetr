import { Moment } from "moment";
import { mustParseToMoment, parseToMomentMaybe } from 'util/parseToMoment';

export default class FundingSchedule {
  fundingScheduleId: number;
  bankAccountId: number;
  name: string;
  description?: string;
  rule: string;
  lastOccurrence?: Moment;
  nextOccurrence: Moment;

  constructor(data?: Partial<FundingSchedule>) {
    if (data) {
      Object.assign(this, {
        ...data,
        lastOccurrence: parseToMomentMaybe(data.lastOccurrence),
        nextOccurrence: mustParseToMoment(data.nextOccurrence),
      });
    }
  }
}
