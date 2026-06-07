import { ID, idPrefix } from '@monetr/interface/models/ID';
import Link from '@monetr/interface/models/Link';
import LunchFlowBankAccount from '@monetr/interface/models/LunchFlowBankAccount';
import PlaidBankAccount from '@monetr/interface/models/PlaidBankAccount';
import { WithJsonValues } from '@monetr/interface/util/json';
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

  bankAccountId: ID<BankAccount>;
  linkId: ID<Link>;
  lunchFlowBankAccountId: ID<LunchFlowBankAccount> | null;
  mask: string | null;
  name: string;
  originalName: string;
  status: BankAccountStatus;
  accountType: BankAccountType;
  accountSubType: BankAccountSubType;
  currency: string;
  // Don't use these fields directly except when creating!
  currentBalance: number;
  availableBalance: number;
  limitBalance: number | null;
  lastUpdated: Date;
  createdAt: Date;
  createdBy: string;
  deletedAt: Date | null;

  plaidBankAccount: PlaidBankAccount | null;
  lunchFlowBankAccount: LunchFlowBankAccount | null;

  constructor(data: WithJsonValues<BankAccount>) {
    this.bankAccountId = ID.from(data.bankAccountId);
    this.linkId = ID.from(data.linkId);
    this.lunchFlowBankAccountId = data.lunchFlowBankAccountId ? ID.from(data.lunchFlowBankAccountId) : null;
    this.mask = data.mask ?? null;

    this.plaidBankAccount = data.plaidBankAccount ? new PlaidBankAccount(data.plaidBankAccount) : null;
    this.lunchFlowBankAccount = data.lunchFlowBankAccount ? new LunchFlowBankAccount(data.lunchFlowBankAccount) : null;

    if (data) {
      Object.assign(this, {
        ...data,
        plaidBankAccount: data?.plaidBankAccount && new PlaidBankAccount(data.plaidBankAccount),
        lunchFlowBankAccount: data?.lunchFlowBankAccount && new LunchFlowBankAccount(data.lunchFlowBankAccount),
        lastUpdated: parseDate(data?.lastUpdated),
        createdAt: parseDate(data?.createdAt),
        deletedAt: parseDate(data?.deletedAt),
      });
    }
  }
}
