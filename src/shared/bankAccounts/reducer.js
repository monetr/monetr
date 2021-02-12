import BankAccountsState from "./state";
import {
  CHANGE_BANK_ACCOUNT,
  FETCH_BANK_ACCOUNT_FAILURE,
  FETCH_BANK_ACCOUNT_SUCCESS,
  FETCH_BANK_ACCOUNTS_REQUEST
} from "./actions";
import {LOGOUT} from "../authentication/actions";


export default function reducer(state = new BankAccountsState(), action) {
  switch (action.type) {
    case FETCH_BANK_ACCOUNTS_REQUEST:
      return state.merge({
        loading: true,
      });
    case FETCH_BANK_ACCOUNT_FAILURE:
      return state.merge({
        loading: false,
      });
    case FETCH_BANK_ACCOUNT_SUCCESS:
      return state.merge({
        loaded: true,
        loading: false,
        bankAccounts: action.payload,
      });
    case CHANGE_BANK_ACCOUNT:
      return state.merge({
        selectedBankAccountId: action.payload,
      });
    case LOGOUT:
      return new BankAccountsState();
    default:
      return state;
  }
}
