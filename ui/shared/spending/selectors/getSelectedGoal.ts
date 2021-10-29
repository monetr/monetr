import Spending from 'models/Spending';
import { Map } from 'immutable';
import { createSelector } from 'reselect';
import { getGoals } from 'shared/spending/selectors/getGoals';

const getSelectedGoalId = state => state.spending.selectedGoalId;

export const getSelectedGoal = createSelector<any, any, Spending | null>(
  [getSelectedGoalId, getGoals],
  (selectedGoalId: number | null, goals: Map<number, Spending>) => {
    if (!selectedGoalId) {
      return null;
    }

    return goals.get(selectedGoalId, null);
  }
)
