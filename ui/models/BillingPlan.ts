export enum BillingInterval {
  Monthly = 'month'
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

  constructor(data: Partial<BillingPlan>) {
    if (data) {
      Object.assign(this, data);
    }
  }
}
