import { Map } from 'immutable';
import Spending from 'models/Spending';
import { createSelector } from 'reselect';
import { getExpenses } from 'shared/spending/selectors/getExpenses';

export const getExpenseIds = createSelector<any, any, number[]>(
  [getExpenses],
  (expenses: Map<number, Spending>) => expenses.sortBy(item => item.name.toLowerCase()).keySeq().toArray(),
);
