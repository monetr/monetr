import LunchFlowBankAccount from '@monetr/interface/models/LunchFlowBankAccount';
import PlaidBankAccount from '@monetr/interface/models/PlaidBankAccount';
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
  bankAccountId: string;
  linkId: string;
  lunchFlowBankAccountId?: string;
  mask?: string;
  name: string;
  originalName: string;
  status: BankAccountStatus;
  accountType: BankAccountType;
  accountSubType: BankAccountSubType;
  currency: string;
  // Don't use these fields directly except when creating!
  currentBalance: number;
  availableBalance: number;
  limitBalance?: number;
  lastUpdated: Date;
  createdAt: Date;
  createdBy: string;
  deletedAt?: Date;

  plaidBankAccount: PlaidBankAccount | null;
  lunchFlowBankAccount: LunchFlowBankAccount | null;

  constructor(data?: Partial<BankAccount>) {
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
