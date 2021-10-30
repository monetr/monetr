import FundingSchedule from 'models/FundingSchedule';
import { createSelector } from 'reselect';
import { getFundingSchedules } from 'shared/fundingSchedules/selectors/getFundingSchedules';

export const getFundingScheduleById = (fundingScheduleId: number) => createSelector<any, any, FundingSchedule>(
  [getFundingSchedules],
  fundingSchedules => fundingSchedules.get(fundingScheduleId, new FundingSchedule()),
);
