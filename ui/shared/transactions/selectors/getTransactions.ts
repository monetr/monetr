import { OrderedMap } from 'immutable';
import Transaction from 'models/Transaction';
import { createSelector } from 'reselect';
import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';

const getTransactionsByBankAccount = state => state.transactions.items;

export const getTransactions = createSelector(
  [getSelectedBankAccountId, getTransactionsByBankAccount],
  (selectedBankAccountId: number, byBankAccount: Map<number, OrderedMap<number, Transaction>>) => {
    return byBankAccount.get(selectedBankAccountId) ?? OrderedMap<number, Transaction>();
  },
);
