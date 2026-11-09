import { ID, idPrefix } from '@monetr/interface/models/ID';
import type SimpleFinLink from '@monetr/interface/models/SimpleFinLink';
import type User from '@monetr/interface/models/User';
import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export enum SimpleFinBankAccountStatus {
  Active = 'active',
  Inactive = 'inactive',
}

export default class SimpleFinBankAccount {
  readonly [idPrefix] = 'sbac';

  simpleFinBankAccountId: ID<SimpleFinBankAccount>;
  simpleFinLinkId: ID<SimpleFinLink>;
  simpleFinId: string;
  connectionId: string;
  name: string;
  institutionName: string;
  currency: string;
  status: SimpleFinBankAccountStatus;
  currentBalance: number;
  availableBalance: number;
  createdAt: Date;
  createdBy: ID<User>;
  updatedAt: Date;
  deletedAt: Date | null;

  constructor(data: WithJsonValues<SimpleFinBankAccount>) {
    this.simpleFinBankAccountId = ID.from(data.simpleFinBankAccountId);
    this.simpleFinLinkId = ID.from(data.simpleFinLinkId);
    this.simpleFinId = data.simpleFinId;
    this.connectionId = data.connectionId;
    this.name = data.name;
    this.institutionName = data.institutionName;
    this.currency = data.currency;
    this.status = data.status;
    this.currentBalance = data.currentBalance;
    this.availableBalance = data.availableBalance;
    this.createdAt = parseDate(data.createdAt);
    this.createdBy = ID.from(data.createdBy);
    this.updatedAt = parseDate(data.updatedAt);
    this.deletedAt = parseDate(data.deletedAt);
  }
}
