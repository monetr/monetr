import { parseJSON } from 'date-fns';

import PlaidBankAccount from '@monetr/interface/models/PlaidBankAccount';
import { formatAmount } from '@monetr/interface/util/amounts';

export type BankAccountStatus = 'unknown' | 'active' | 'inactive';

export default class BankAccount {
  bankAccountId: string;
  linkId: string;
  availableBalance: number;
  currentBalance: number;
  mask?: string;
  name: string;
  originalName: string;
  status: BankAccountStatus;
  accountType: string;
  accountSubType: string;
  lastUpdated: Date;
  createdAt: Date;
  createdBy: number;

  plaidBankAccount: PlaidBankAccount | null;

  constructor(data?: Partial<BankAccount>) {
    if (data) {
      Object.assign(this, {
        ...data,
        plaidBankAccount: data?.plaidBankAccount && new PlaidBankAccount(data.plaidBankAccount),
        lastUpdated: data?.lastUpdated && parseJSON(data.lastUpdated),
        createdAt: data?.createdAt && parseJSON(data.createdAt),
      });
    }
  }

  getAvailableBalanceString() {
    return formatAmount(this.availableBalance);
  }

  getCurrentBalanceString() {
    return formatAmount(this.currentBalance);
  }
}
