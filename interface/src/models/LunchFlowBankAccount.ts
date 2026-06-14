import { ID, idPrefix } from '@monetr/interface/models/ID';
import type LunchFlowLink from '@monetr/interface/models/LunchFlowLink';
import type User from '@monetr/interface/models/User';
import type { WithJsonValues } from '@monetr/interface/util/json';
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
  readonly [idPrefix] = 'lbac';

  lunchFlowBankAccountId: ID<LunchFlowBankAccount>;
  lunchFlowLinkId: ID<LunchFlowLink>;
  lunchFlowId: string;
  lunchFlowStatus: LunchFlowBankAccountExternalStatus;
  name: string;
  institutionName: string;
  provider: string;
  currency: string;
  status: LunchFlowBankAccountStatus;
  currentBalance: number;
  updatedAt: Date;
  createdAt: Date;
  deletedAt: Date | null;
  createdBy: ID<User>;

  constructor(data: WithJsonValues<LunchFlowBankAccount>) {
    this.lunchFlowBankAccountId = ID.from(data.lunchFlowBankAccountId);
    this.lunchFlowLinkId = ID.from(data.lunchFlowLinkId);
    this.lunchFlowId = data.lunchFlowId;
    this.lunchFlowStatus = data.lunchFlowStatus;
    this.name = data.name;
    this.institutionName = data.institutionName;
    this.provider = data.provider;
    this.currency = data.currency;
    this.status = data.status;
    this.currentBalance = data.currentBalance;
    this.updatedAt = parseDate(data.updatedAt);
    this.createdAt = parseDate(data.createdAt);
    this.deletedAt = parseDate(data.deletedAt);
    this.createdBy = ID.from(data.createdBy);
  }
}
