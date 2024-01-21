import { parseJSON } from 'date-fns';

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
        expirationDate: data?.expirationDate && parseJSON(data.expirationDate),
        lastManualSync: data?.lastManualSync && parseJSON(data.lastManualSync),
        lastSuccessfulUpdate: data?.lastSuccessfulUpdate && parseJSON(data.lastSuccessfulUpdate),
        lastAttemptedUpdate: data?.lastAttemptedUpdate && parseJSON(data.lastAttemptedUpdate),
        updatedAt: data?.updatedAt && parseJSON(data.updatedAt),
        createdAt: data?.createdAt && parseJSON(data.createdAt),
      });
    }
  }
}
