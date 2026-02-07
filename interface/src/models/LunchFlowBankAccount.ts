import parseDate from '@monetr/interface/util/parseDate';

export enum LunchFlowBankAccountStatus {
  Active = 'active',
  Inactive = 'inactive',
  Error = 'error',
}

export enum LunchFlowBankAccountExternalStatus {
  Active = 'ACTIVE',
  Disconnected = 'DISCONNECTED',
  Error = 'ERROR',
}

export default class LunchFlowBankAccount {
  lunchFlowBankAccountId: string;
  lunchFlowLinkId: string;
  LunchFlowId: string;
  lunchFlowStatus: LunchFlowBankAccountExternalStatus;
  name: string;
  institutionName: string;
  provider: string;
  currency: string;
  status: LunchFlowBankAccountStatus;
  currentBalance: number;
  updatedAt: Date;
  createdAt: Date;
  deletedAt?: Date;
  createdBy: string;

  constructor(data?: Partial<LunchFlowBankAccount>) {
    if (data) {
      Object.assign(this, {
        ...data,
        updatedAt: parseDate(data?.updatedAt),
        createdAt: parseDate(data?.createdAt),
        deletedAt: parseDate(data?.deletedAt),
      });
    }
  }
}
