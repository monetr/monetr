import { ID, idPrefix } from '@monetr/interface/models/ID';
import type User from '@monetr/interface/models/User';
import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export enum SimpleFinLinkStatus {
  Pending = 'pending',
  Active = 'active',
  Deactivated = 'deactivated',
  Error = 'error',
}

export interface SimpleFinLinkError {
  code: string;
  message: string;
  connectionId: string | null;
  simpleFinId: string | null;
}

export default class SimpleFinLink {
  readonly [idPrefix] = 'sfx';

  simpleFinLinkId: ID<SimpleFinLink>;
  name: string;
  apiUrl: string;
  status: SimpleFinLinkStatus;
  lastErrors: Array<SimpleFinLinkError>;
  lastManualSync: Date | null;
  lastSuccessfulUpdate: Date | null;
  lastAttemptedUpdate: Date | null;
  createdAt: Date;
  createdBy: ID<User>;
  updatedAt: Date;
  deletedAt: Date | null;

  constructor(data: WithJsonValues<SimpleFinLink>) {
    this.simpleFinLinkId = ID.from(data.simpleFinLinkId);
    this.name = data.name;
    this.apiUrl = data.apiUrl;
    this.status = data.status;
    this.lastErrors = data.lastErrors ?? [];
    this.lastManualSync = parseDate(data.lastManualSync);
    this.lastSuccessfulUpdate = parseDate(data.lastSuccessfulUpdate);
    this.lastAttemptedUpdate = parseDate(data.lastAttemptedUpdate);
    this.createdAt = parseDate(data.createdAt);
    this.createdBy = ID.from(data.createdBy);
    this.updatedAt = parseDate(data.updatedAt);
    this.deletedAt = parseDate(data.deletedAt);
  }
}
