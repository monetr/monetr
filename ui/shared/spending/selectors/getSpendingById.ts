import { Map } from 'immutable';
import Spending from 'models/Spending';
import { createSelector } from 'reselect';
import { getSpending } from 'shared/spending/selectors/getSpending';

export const getSpendingById = (expenseId?: number) => createSelector<any, any, Spending | null>(
  getSpending,
  (expenses: Map<number, Spending>) => {
    if (!expenseId) {
      return null;
    }

    return expenses.get(expenseId, null);
  },
);
