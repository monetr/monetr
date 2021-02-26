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
      return {
        ...state,
        loaded: true,
        loading: false,
        items: state.items.merge(action.payload)
      }
    case LOGOUT:
      return new BankAccountsState();
    default:
      return state;
  }
}
