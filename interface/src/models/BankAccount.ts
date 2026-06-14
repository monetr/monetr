import { ID, idPrefix } from '@monetr/interface/models/ID';
import type Link from '@monetr/interface/models/Link';
import LunchFlowBankAccount from '@monetr/interface/models/LunchFlowBankAccount';
import PlaidBankAccount from '@monetr/interface/models/PlaidBankAccount';
import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export type BankAccountStatus = 'unknown' | 'active' | 'inactive';

export enum BankAccountType {
  Depository = 'depository',
  Credit = 'credit',
  Loan = 'loan',
  Investment = 'investment',
  Other = 'other',
}

export enum BankAccountSubType {
  Checking = 'checking',
  Savings = 'savings',
  HSA = 'hsa',
  CD = 'cd',
  MoneyMarket = 'money market',
  PayPal = 'paypal',
  Prepaid = 'prepaid',
  CashManagement = 'cash management',
  EBT = 'ebt',
  CreditCard = 'credit card',
  Auto = 'auto',
  Other = 'other',
}

export default class BankAccount {
  readonly [idPrefix] = 'bac';

  readonly bankAccountId: ID<BankAccount>;
  readonly linkId: ID<Link>;
  readonly lunchFlowBankAccountId: ID<LunchFlowBankAccount> | null;
  mask: string | null;
  name: string;
  readonly originalName: string;
  readonly status: BankAccountStatus;
  accountType: BankAccountType;
  accountSubType: BankAccountSubType;
  currency: string;
  // Don't use these fields directly except when creating!
  currentBalance: number;
  availableBalance: number;
  limitBalance: number | null;
  readonly lastUpdated: Date;
  readonly createdAt: Date;
  readonly createdBy: string;
  readonly deletedAt: Date | null;

  readonly plaidBankAccount: PlaidBankAccount | null;
  readonly lunchFlowBankAccount: LunchFlowBankAccount | null;

  constructor(data: WithJsonValues<BankAccount>) {
    this.bankAccountId = ID.from(data.bankAccountId);
    this.linkId = ID.from(data.linkId);
    this.lunchFlowBankAccountId = data.lunchFlowBankAccountId ? ID.from(data.lunchFlowBankAccountId) : null;
    this.mask = data.mask ?? null;
    this.name = data.name;
    this.originalName = data.originalName;
    this.status = data.status;
    this.accountType = data.accountType;
    this.accountSubType = data.accountSubType;
    this.currency = data.currency;
    this.currentBalance = data.currentBalance;
    this.availableBalance = data.availableBalance;
    this.limitBalance = data.limitBalance ?? null;
    this.lastUpdated = parseDate(data.lastUpdated);
    this.createdAt = parseDate(data.createdAt);
    this.createdBy = data.createdBy;
    this.deletedAt = parseDate(data.deletedAt);
    this.plaidBankAccount = data.plaidBankAccount ? new PlaidBankAccount(data.plaidBankAccount) : null;
    this.lunchFlowBankAccount = data.lunchFlowBankAccount ? new LunchFlowBankAccount(data.lunchFlowBankAccount) : null;
  }
}
