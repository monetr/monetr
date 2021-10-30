import Spending from 'models/Spending';
import { Map } from 'immutable';
import { createSelector } from 'reselect';
import { getGoals } from 'shared/spending/selectors/getGoals';


export const getGoalIds = createSelector<any, any, number[]>(
  [getGoals],
  (goals: Map<number, Spending>) => goals.keySeq().toArray(),
);
