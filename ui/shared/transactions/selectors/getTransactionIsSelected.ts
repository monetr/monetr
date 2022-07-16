import { createSelector } from 'reselect';

const getSelectedTransactionId = state => state.transactions.selectedTransactionId;

export const getTransactionIsSelected = (transactionId: number) => createSelector<any, any, boolean>(
  [getSelectedTransactionId],
  selectedTransactionId => transactionId === selectedTransactionId
);
