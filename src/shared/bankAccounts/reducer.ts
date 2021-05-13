import { LOGOUT } from "shared/authentication/actions";
import {
  BankAccountActions,
  CHANGE_BANK_ACCOUNT,
  FETCH_BANK_ACCOUNTS_FAILURE,
  FETCH_BANK_ACCOUNTS_REQUEST,
  FETCH_BANK_ACCOUNTS_SUCCESS
} from "shared/bankAccounts/actions";
import BankAccountsState from "shared/bankAccounts/state";

export default function reducer(state: BankAccountsState = new BankAccountsState(), action: BankAccountActions) {
  switch (action.type) {
    case CHANGE_BANK_ACCOUNT:
      return {
        ...state,
        selectedBankAccountId: action.bankAccountId,
      }
    case FETCH_BANK_ACCOUNTS_REQUEST:
      return {
        ...state,
        loading: true,
      };
    case FETCH_BANK_ACCOUNTS_FAILURE:
      return {
        ...state,
        loading: false,
      };
    case FETCH_BANK_ACCOUNTS_SUCCESS:
      // If there is a bank account selected in redux, use that bank account.
      let selectedBankAccountId = state.selectedBankAccountId
        // If there is not, check and see if there is one in local storage that is valid.
        || +window.localStorage.getItem('selectedBankAccountId');

      const allBankAccounts = state.items.merge(action.payload)

      // If there is a bankAccountId selected, but its not in our list of bank accounts -> then default to the first
      // bank account we have.
      if (selectedBankAccountId && !allBankAccounts.has(selectedBankAccountId)) {
        selectedBankAccountId = allBankAccounts.first(null)?.bankAccountId;
        // Remove the local storage item since it's not considered accurate.
        window.localStorage.removeItem('selectedBankAccountId');
      }

      return {
        ...state,
        loaded: true,
        loading: false,
        selectedBankAccountId: selectedBankAccountId,
        items: state.items.merge(action.payload)
      }
    case LOGOUT:
      return new BankAccountsState();
    default:
      return state;
  }
}
