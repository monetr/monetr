import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export enum PlaidLinkStatus {
  Unknown = 'unknown',
  Pending = 'pending',
  Setup = 'setup',
  Error = 'error',
  PendingExpiration = 'pending_expiration',
  Revoked = 'revoked',
  Deactivated = 'deactivated',
}

export default class PlaidLink {
  products: Array<string>;
  status: PlaidLinkStatus;
  errorCode?: string;
  expirationDate?: Date;
  newAccountsAvailable: boolean;
  institutionId: string;
  institutionName: string;
  lastManualSync?: Date;
  lastSuccessfulUpdate?: Date;
  lastAttemptedUpdate?: Date;
  updatedAt: Date;
  createdAt: Date;
  createdBy: string;

  constructor(data: WithJsonValues<PlaidLink>) {
    this.products = data.products;
    this.status = data.status;
    this.errorCode = data.errorCode;
    // These dates are all optional so we coalesce the null that parseDate returns back into undefined to match the
    // field types.
    this.expirationDate = parseDate(data.expirationDate) ?? undefined;
    this.newAccountsAvailable = data.newAccountsAvailable;
    this.institutionId = data.institutionId;
    this.institutionName = data.institutionName;
    this.lastManualSync = parseDate(data.lastManualSync) ?? undefined;
    this.lastSuccessfulUpdate = parseDate(data.lastSuccessfulUpdate) ?? undefined;
    this.lastAttemptedUpdate = parseDate(data.lastAttemptedUpdate) ?? undefined;
    this.updatedAt = parseDate(data.updatedAt);
    this.createdAt = parseDate(data.createdAt);
    this.createdBy = data.createdBy;
  }
}
