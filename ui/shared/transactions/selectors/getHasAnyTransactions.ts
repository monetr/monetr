import Transaction from "models/Transaction";
import { Map, OrderedMap } from "immutable";
import { createSelector } from "reselect";
import { getSelectedBankAccountId } from "shared/bankAccounts/selectors/getSelectedBankAccountId";

const getTransactionsByBankAccount = state => state.transactions.items;

// getHasAnyTransactions will return true or false indicating whether or not the currently selected bank account has
// any transactions currently retrieved and stored in redux. It is not an indicator that the account does or does not
// actually have transactions, only whether or not they have been retrieved.
export const getHasAnyTransactions = createSelector(
  [getTransactionsByBankAccount, getSelectedBankAccountId],
  (transactions: Map<number, OrderedMap<number, Transaction>>, selectedBankAccountId) => {
    return !transactions.get(selectedBankAccountId, OrderedMap<number, Transaction>()).isEmpty();
  }
);
