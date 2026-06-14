import type { WithJsonValues } from '@monetr/interface/util/json';

export enum BillingInterval {
  Monthly = 'month',
}

export default class BillingPlan {
  id: string;
  name: string;
  description: string;
  unitPrice: number;
  interval: BillingInterval;
  intervalCount: number;
  freeTrialDays: number;
  active: boolean;

  constructor(data: WithJsonValues<BillingPlan>) {
    this.id = data.id;
    this.name = data.name;
    this.description = data.description;
    this.unitPrice = data.unitPrice;
    this.interval = data.interval;
    this.intervalCount = data.intervalCount;
    this.freeTrialDays = data.freeTrialDays;
    this.active = data.active;
  }
}
