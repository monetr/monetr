import { createSelector } from "reselect";
import { getSelectedBankAccountId } from "shared/bankAccounts/selectors/getSelectedBankAccountId";
import Transaction from "models/Transaction";
import { OrderedMap } from "immutable";

const getTransactionsByBankAccount = state => state.transactions.items;

export const getTransactionIds = createSelector(
  [getSelectedBankAccountId, getTransactionsByBankAccount],
  (selectedBankAccountId: number, byBankAccount: Map<number, OrderedMap<number, Transaction>>) => {
    return (byBankAccount.get(selectedBankAccountId) ?? OrderedMap<number, Transaction>()).keySeq().toArray();
  },
);
