import { Map, Record } from 'immutable';
import BankAccount from "data/BankAccount";

export default class BankAccountsState extends Record({
  items: Map<number, BankAccount>(),
  loaded: false,
  loading: false,
}) {
  items: Map<number, BankAccount>;
  loaded: boolean;
  loading: boolean;
  selectedBankAccountId?: number;
}
