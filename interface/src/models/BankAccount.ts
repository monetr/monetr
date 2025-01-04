
import PlaidBankAccount from '@monetr/interface/models/PlaidBankAccount';
import { formatAmount } from '@monetr/interface/util/amounts';
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
  availableBalance: number;
  currentBalance: number;
  limitBalance: number;
  mask?: string;
  name: string;
  originalName: string;
  status: BankAccountStatus;
  accountType: BankAccountType;
  accountSubType: BankAccountSubType;
  lastUpdated: Date;
  createdAt: Date;
  createdBy: string;

  plaidBankAccount: PlaidBankAccount | null;

  constructor(data?: Partial<BankAccount>) {
    if (data) {
      Object.assign(this, {
        ...data,
        plaidBankAccount: data?.plaidBankAccount && new PlaidBankAccount(data.plaidBankAccount),
        lastUpdated: parseDate(data?.lastUpdated),
        createdAt: parseDate(data?.createdAt),
      });
    }
  }

  getAvailableBalanceString() {
    return formatAmount(this.availableBalance);
  }

  getCurrentBalanceString() {
    return formatAmount(this.currentBalance);
  }

  getLimitBalanceString() {
    return formatAmount(this.limitBalance);
  }
}
