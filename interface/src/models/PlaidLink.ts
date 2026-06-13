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
  errorCode: string | null;
  expirationDate: Date | null;
  newAccountsAvailable: boolean;
  institutionId: string;
  institutionName: string;
  lastManualSync: Date | null;
  lastSuccessfulUpdate: Date | null;
  lastAttemptedUpdate: Date | null;
  updatedAt: Date;
  createdAt: Date;
  createdBy: string;

  constructor(data: WithJsonValues<PlaidLink>) {
    this.products = data.products;
    this.status = data.status;
    this.errorCode = data.errorCode ?? null;
    // These dates are all nullable, parseDate already hands back null when the API omits them.
    this.expirationDate = parseDate(data.expirationDate);
    this.newAccountsAvailable = data.newAccountsAvailable;
    this.institutionId = data.institutionId;
    this.institutionName = data.institutionName;
    this.lastManualSync = parseDate(data.lastManualSync);
    this.lastSuccessfulUpdate = parseDate(data.lastSuccessfulUpdate);
    this.lastAttemptedUpdate = parseDate(data.lastAttemptedUpdate);
    this.updatedAt = parseDate(data.updatedAt);
    this.createdAt = parseDate(data.createdAt);
    this.createdBy = data.createdBy;
  }
}
