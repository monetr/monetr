import { Map } from 'immutable';
import BankAccount from "models/BankAccount";

export default class BankAccountsState {
  constructor() {
    this.items = Map<number, BankAccount>();
  }

  items: Map<number, BankAccount>;
  loaded: boolean;
  loading: boolean;
  selectedBankAccountId?: number;
}
