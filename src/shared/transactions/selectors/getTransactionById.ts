import { createSelector } from "reselect";
import { getTransactions } from "shared/transactions/selectors/getTransactions";


export const getTransactionById = (transactionId: number) => createSelector(
  [getTransactions],
  transactions => transactions.get(transactionId, null),
);
