export interface BankAccountFields {
  bankAccountId: number;
  linkId: number;
  availableBalance: number;
  currentBalance: number;
  mask?: string;
  name: string;
  plaidName?: string;
  plaidOfficialName?: string;
  type: string;
  subType: string;
}

export default class BankAccount implements BankAccountFields {
  bankAccountId: number;
  linkId: number;
  availableBalance: number;
  currentBalance: number;
  mask?: string;
  name: string;
  plaidName?: string;
  plaidOfficialName?: string;
  type: string;
  subType: string;

  constructor(data: BankAccountFields) {
    Object.assign(this, data)
  }

  getAvailableBalanceString() {
    return `$${ (this.availableBalance / 100).toFixed(2) }`;
  }

  getCurrentBalanceString() {
    return `$${ (this.currentBalance / 100).toFixed(2) }`;
  }
}
