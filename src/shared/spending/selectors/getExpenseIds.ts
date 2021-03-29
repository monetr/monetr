import { createSelector } from 'reselect';
import { getExpenses } from 'shared/spending/selectors/getExpenses';

export const getExpenseIds = createSelector<any, any, number[]>(
  [getExpenses],
  expenses => expenses.keySeq().toArray(),
);
