import { ID, idPrefix } from '@monetr/interface/models/ID';
import type User from '@monetr/interface/models/User';
import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export enum LunchFlowLinkStatus {
  Pending = 'pending',
  Active = 'active',
  Deactivated = 'deactivated',
  Error = 'error',
}

export default class LunchFlowLink {
  readonly [idPrefix] = 'lfx';

  lunchFlowLinkId: ID<LunchFlowLink>;
  name: string;
  apiUrl: string;
  status: LunchFlowLinkStatus;
  lastManualSync: Date | null;
  lastSuccessfulUpdate: Date | null;
  lastAttemptedUpdate: Date | null;
  updatedAt: Date;
  createdAt: Date;
  deletedAt: Date | null;
  createdBy: ID<User>;

  constructor(data: WithJsonValues<LunchFlowLink>) {
    this.lunchFlowLinkId = ID.from(data.lunchFlowLinkId);
    this.name = data.name;
    this.apiUrl = data.apiUrl;
    this.status = data.status;
    this.lastManualSync = parseDate(data.lastManualSync);
    this.lastSuccessfulUpdate = parseDate(data.lastSuccessfulUpdate);
    this.lastAttemptedUpdate = parseDate(data.lastAttemptedUpdate);
    this.updatedAt = parseDate(data.updatedAt);
    this.createdAt = parseDate(data.createdAt);
    this.deletedAt = parseDate(data.deletedAt);
    this.createdBy = ID.from(data.createdBy);
  }
}
