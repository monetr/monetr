import { Map } from 'immutable';
import Spending from 'models/Spending';
import { createSelector } from 'reselect';
import { getSpending } from 'shared/spending/selectors/getSpending';

export const getExpenses = createSelector<any, any, Map<number, Spending>>(
  [getSpending],
  (spending: Map<number, Spending>) => {
    return spending.filter(spend => spend.getIsExpense());
  }
);
