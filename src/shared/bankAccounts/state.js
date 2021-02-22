import {Map, Record} from 'immutable';

export default class BankAccountsState extends Record({
  items: new Map(),
  loaded: false,
  loading: false,
  // When the user selects a bank account that they want to view, this value is changed.
  selectedBankAccountId: null,
}) {

}
