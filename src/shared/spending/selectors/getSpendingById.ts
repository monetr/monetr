import { createSelector } from "reselect";
import { getSpending } from "shared/spending/selectors/getSpending";
import Spending from "data/Spending";


export const getSpendingById = (expenseId?: number) => createSelector<any, any, Spending|null>(
  [getSpending],
  expenses => {
    if (!expenseId) {
      return null;
    }

    return expenses.get(expenseId, null)
  },
);
