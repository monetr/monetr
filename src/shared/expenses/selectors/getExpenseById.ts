import { createSelector } from "reselect";
import { getExpenses } from "shared/expenses/selectors/getExpenses";
import Expense from "data/Expense";


export const getExpenseById = (expenseId?: number) => createSelector<any, any, Expense|null>(
  [getExpenses],
  expenses => {
    if (!expenseId) {
      return null;
    }

    return expenses.get(expenseId, null)
  },
);
