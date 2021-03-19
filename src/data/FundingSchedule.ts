import { Moment } from "moment";

export interface FundingScheduleFields {
  fundingScheduleId: number;
  bankAccountId: number;
  name: string;
  description?: string;
  rule: string;
  lastOccurrence?: Moment;
  nextOccurrence: Moment;
}

export default class FundingSchedule implements FundingScheduleFields {
  fundingScheduleId: number;
  bankAccountId: number;
  name: string;
  description?: string;
  rule: string;
  lastOccurrence?: Moment;
  nextOccurrence: Moment;

  constructor(data?: FundingScheduleFields) {
    if (data) {
      Object.assign(this, data);
    }
  }
}
