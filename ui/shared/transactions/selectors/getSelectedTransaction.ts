import { OrderedMap } from 'immutable';
import Transaction from 'models/Transaction';
import { createSelector } from 'reselect';
import { getTransactions } from 'shared/transactions/selectors/getTransactions';

const getSelectedTransactionId = state => state.transactions.selectedTransactionId;

export const getSelectedTransaction = createSelector<any, any, Transaction|null>(
  [getSelectedTransactionId, getTransactions],
  (selectedTransactionId: number|null, transactions: OrderedMap<number, Transaction>) => {
    if (!selectedTransactionId) {
      return null;
    }

    return transactions.get(selectedTransactionId, null);
  },
);
