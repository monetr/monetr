import Spending from 'models/Spending';
import { createSelector } from 'reselect';
import { getSpending } from 'shared/spending/selectors/getSpending';
import { Map } from 'immutable';

export const getFundingScheduleContribution = (fundingScheduleId: number) => createSelector<any, any, number>(
  [getSpending],
  (spending: Map<number, Spending>) => spending.reduce((total: number, item: Spending) => {
    return total + (item.fundingScheduleId === fundingScheduleId ? item.nextContributionAmount : 0);
  }, 0),
);
