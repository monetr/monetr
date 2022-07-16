import BankAccount from 'models/BankAccount';
import { Logout } from 'shared/authentication/actions';
import {
  BankAccountActions,
  CHANGE_BANK_ACCOUNT,
  FETCH_BANK_ACCOUNTS_FAILURE,
  FETCH_BANK_ACCOUNTS_REQUEST,
  FETCH_BANK_ACCOUNTS_SUCCESS,
} from 'shared/bankAccounts/actions';
import BankAccountsState from 'shared/bankAccounts/state';
import { RemoveLink } from 'shared/links/actions';

export default function reducer(state: BankAccountsState = new BankAccountsState(), action: BankAccountActions): BankAccountsState {
  switch (action.type) {
    case CHANGE_BANK_ACCOUNT:
      return {
        ...state,
        selectedBankAccountId: action.payload,
      };
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

      const allBankAccounts = state.items.merge(action.payload);

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
        items: state.items.merge(action.payload),
      };
    case RemoveLink.Success:
      // This is a bit goofy. Basically when we remove a link we are returned the link itself, and all of the bank
      // accounts associated with that link. We are basically doing a reverse intersection here (read exclusion) to
      // remove the bank accounts that would be removed as part of this link being removed.
      const newBankAccountsSet = state.items.filter((_: BankAccount, bankAccountId: number): boolean => {
        return !action.payload.bankAccounts.has(bankAccountId);
      });
      return {
        ...state,
        loading: false,
        selectedBankAccountId: !newBankAccountsSet.has(state.selectedBankAccountId) ? newBankAccountsSet.first()?.bankAccountId || null : state.selectedBankAccountId,
        items: newBankAccountsSet,
      };
    case Logout.Success:
      return new BankAccountsState();
    default:
      return state;
  }
}
