import { createSelector } from 'reselect';

const getSelectedGoalId = state => state.spending.selectedGoalId;

export const getGoalIsSelected = (goalId: number) => createSelector<any, any, boolean>(
  [getSelectedGoalId],
  selectedGoalId => goalId === selectedGoalId,
);
