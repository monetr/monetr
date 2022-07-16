import { List } from 'immutable';
import { createSelector } from 'reselect';
import { getFundingSchedules } from 'shared/fundingSchedules/selectors/getFundingSchedules';

export const getFundingScheduleIds = createSelector<any, any, List<number>>(
  [getFundingSchedules],
  fundingSchedules => {
    return fundingSchedules.keySeq().toList();
  }
);
