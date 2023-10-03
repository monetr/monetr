
export enum Feature {
  ManualBudgeting = 'ManualBudgeting',
  LinkedBudgeting = 'LinkedBudgeting',
}

export enum SubscriptionStatus {
  Active = 'active',
}

export default class Subscription {
  subscriptionId: number;
  ownedByUserId: number;
  features: Feature[];
  status: SubscriptionStatus;
  trialStart: Date | null;
  trialEnd: Date | null;

  constructor(data: Partial<Subscription>) {
    if (data) {
      Object.assign(this, data);
    }
  }

  hasFeature(feature: Feature): boolean {
    return this.features.includes(feature);
  }
}
