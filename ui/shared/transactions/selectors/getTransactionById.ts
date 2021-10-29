import { createSelector } from "reselect";
import { getTransactions } from "shared/transactions/selectors/getTransactions";
import Transaction from "models/Transaction";


export const getTransactionById = (transactionId: number) => createSelector<any, any, Transaction|null>(
  [getTransactions],
  transactions => transactions.get(transactionId, null),
);
