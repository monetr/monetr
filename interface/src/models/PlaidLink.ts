import parseDate from '@monetr/interface/util/parseDate';

export enum PlaidLinkStatus {
  Unknown = 0,
  Pending = 1,
  Setup = 2,
  Error = 3,
  PendingExpiration = 4,
  Revoked = 5,
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
  createdByUserId: number;

  constructor(data?: Partial<PlaidLink>) {
    if (data) {
      Object.assign(this, {
        ...data,
        expirationDate: parseDate(data?.expirationDate),
        lastManualSync: parseDate(data?.lastManualSync),
        lastSuccessfulUpdate: parseDate(data?.lastSuccessfulUpdate),
        lastAttemptedUpdate: parseDate(data?.lastAttemptedUpdate),
        updatedAt: parseDate(data?.updatedAt),
        createdAt: parseDate(data?.createdAt),
      });
    }
  }
}
