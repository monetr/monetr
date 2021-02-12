import {Record} from "immutable";


export default class BankAccount extends Record({
  bankAccountId: 0,
  linkId: 0,
  availableBalance: 0,
  currentBalance: 0,
  mask: '',
  name: null,
  plaidName: '',
  plaidOfficialName: null,
  type: null,
  subType: null,
}) {
  bankAccountId;
  linkId;
  availableBalance;
  currentBalance;
  mask;
  name;
  plaidName;
  plaidOfficialName;
  type;
  subType;

  getAvailableBalanceString() {
    return `$${(this.availableBalance / 100).toFixed(2)}`;
  }

  getCurrentBalanceString() {
    return `$${(this.currentBalance / 100).toFixed(2)}`;
  }
}
