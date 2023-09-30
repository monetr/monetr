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
    return `$${ (this.availableBalance / 100).toFixed(2) }`;
  }

  getCurrentBalanceString() {
    return `$${ (this.currentBalance / 100).toFixed(2) }`;
  }
}
