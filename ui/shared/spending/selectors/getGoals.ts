import Spending from "models/Spending";
import { Map } from 'immutable';
import { createSelector } from "reselect";
import { getSpending } from "shared/spending/selectors/getSpending";

export const getGoals = createSelector<any, any, Map<number, Spending>>(
  [getSpending],
  (spending: Map<number, Spending>) => {
    return spending.filter(spend => spend.getIsGoal());
  }
);
