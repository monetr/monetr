import { formatAmount } from '@monetr/interface/util/amounts';

export default class BankAccount {
  bankAccountId: number;
  linkId: number;
  availableBalance: number;
  currentBalance: number;
  mask?: string;
  name: string;
  plaidName?: string;
  plaidOfficialName?: string;
  accountType: string;
  accountSubType: string;

  constructor(data?: Partial<BankAccount>) {
    if (data) {
      Object.assign(this, data);
    }
  }

  getAvailableBalanceString() {
    return formatAmount(this.availableBalance);
  }

  getCurrentBalanceString() {
    return formatAmount(this.currentBalance);
  }
}
