
import PlaidBankAccount from '@monetr/interface/models/PlaidBankAccount';
import { AmountType, formatAmount } from '@monetr/interface/util/amounts';
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
  currency: string;
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

  getAvailableBalanceString(locale: string = 'en_US') {
    return formatAmount(this.availableBalance, AmountType.Stored, locale, this.currency);
  }

  getCurrentBalanceString(locale: string = 'en_US') {
    return formatAmount(this.currentBalance, AmountType.Stored, locale, this.currency);
  }

  getLimitBalanceString(locale: string = 'en_US') {
    return formatAmount(this.limitBalance, AmountType.Stored, locale, this.currency);
  }
}
