import { Map } from 'immutable';
import Spending from 'models/Spending';
import { createSelector } from 'reselect';
import { getExpenses } from 'shared/spending/selectors/getExpenses';

const getSelectedExpenseId = state => state.spending.selectedExpenseId;

export const getSelectedExpense = createSelector<any, any, Spending | null>(
  [getSelectedExpenseId, getExpenses],
  (selectedExpenseId: number | null, expenses: Map<number, Spending>) => {
    if (!selectedExpenseId) {
      return null;
    }

    return expenses.get(selectedExpenseId, null);
  }
);
