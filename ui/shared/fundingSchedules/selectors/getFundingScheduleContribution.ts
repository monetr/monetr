import { Map } from 'immutable';
import Spending from 'models/Spending';
import { createSelector } from 'reselect';
import { getSpending } from 'shared/spending/selectors/getSpending';

export const getFundingScheduleContribution = (fundingScheduleId: number) => createSelector<any, any, number>(
  [getSpending],
  (spending: Map<number, Spending>) => spending.reduce((total: number, item: Spending) => {
    // Build a total of all the spending objects contributions who are not paused and are for this funding schedule.
    return total + (item.fundingScheduleId === fundingScheduleId && !item.isPaused ? item.nextContributionAmount : 0);
  }, 0),
);
