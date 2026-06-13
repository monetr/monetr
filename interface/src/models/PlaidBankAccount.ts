import { ID } from '@monetr/interface/models/ID';
import type User from '@monetr/interface/models/User';
import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export default class PlaidBankAccount {
  name: string;
  officialName: string | null;
  mask: string | null;
  availableBalance: number;
  currentBalance: number;
  limitBalance: number | null;
  createdAt: Date;
  createdBy: ID<User>;

  constructor(data: WithJsonValues<PlaidBankAccount>) {
    this.name = data.name;
    this.officialName = data.officialName ?? null;
    this.mask = data.mask ?? null;
    this.availableBalance = data.availableBalance;
    this.currentBalance = data.currentBalance;
    this.limitBalance = data.limitBalance ?? null;
    this.createdAt = parseDate(data.createdAt);
    // The API sends this as `createdBy`, the old field name read `createdByUserId` which the
    // server never actually sends so it was always undefined.
    this.createdBy = ID.from(data.createdBy);
  }
}
