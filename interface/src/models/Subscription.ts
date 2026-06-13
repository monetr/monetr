import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export enum Feature {
  ManualBudgeting = 'ManualBudgeting',
  LinkedBudgeting = 'LinkedBudgeting',
}

export enum SubscriptionStatus {
  Active = 'active',
}

export default class Subscription {
  subscriptionId: string;
  ownedByUserId: string;
  features: Feature[];
  status: SubscriptionStatus;
  trialStart: Date | null;
  trialEnd: Date | null;

  constructor(data: WithJsonValues<Subscription>) {
    this.subscriptionId = data.subscriptionId;
    this.ownedByUserId = data.ownedByUserId;
    this.features = data.features;
    this.status = data.status;
    this.trialStart = parseDate(data.trialStart);
    this.trialEnd = parseDate(data.trialEnd);
  }

  hasFeature(feature: Feature): boolean {
    return this.features.includes(feature);
  }
}
