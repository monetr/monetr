
import PlaidBankAccount from '@monetr/interface/models/PlaidBankAccount';
import parseDate from '@monetr/interface/util/parseDate';

export type BankAccountStatus = 'unknown' | 'active' | 'inactive';

export enum BankAccountType {
	Depository = 'depository',
	Credit     = 'credit',
	Loan       = 'loan',
	Investment = 'investment',
	Other      = 'other',
}

export enum BankAccountSubType {
  Checking       = 'checking',
  Savings        = 'savings',
  HSA            = 'hsa',
  CD             = 'cd',
  MoneyMarket    = 'money market',
  PayPal         = 'paypal',
  Prepaid        = 'prepaid',
  CashManagement = 'cash management',
  EBT            = 'ebt',
  CreditCard     = 'credit card',
  Auto           = 'auto',
  Other          = 'other',
}

export default class BankAccount {
  bankAccountId: string;
  linkId: string;
  mask?: string;
  name: string;
  originalName: string;
  status: BankAccountStatus;
  accountType: BankAccountType;
  accountSubType: BankAccountSubType;
  currency: string;
  lastUpdated: Date;
  createdAt: Date;
  createdBy: string;
  deletedAt?: Date;

  plaidBankAccount: PlaidBankAccount | null;

  constructor(data?: Partial<BankAccount>) {
    if (data) {
      Object.assign(this, {
        ...data,
        plaidBankAccount: data?.plaidBankAccount && new PlaidBankAccount(data.plaidBankAccount),
        lastUpdated: parseDate(data?.lastUpdated),
        createdAt: parseDate(data?.createdAt),
        deletedAt: Boolean(data?.deletedAt) && parseDate(data.deletedAt),
      });
    }
  }
}
