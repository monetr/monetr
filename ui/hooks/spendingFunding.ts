import { useQuery, UseQueryResult } from 'react-query';

import { useFundingSchedulesSink } from 'hooks/fundingSchedules';
import FundingSchedule from 'models/FundingSchedule';
import Spending from 'models/Spending';
import SpendingFunding from 'models/SpendingFunding';

export type SpendingFundingSinkResult =
  { result: Array<SpendingFunding> }
  & UseQueryResult<Array<Partial<SpendingFunding>>>;

export function useSpendingFundingSink(spending: Spending | null): SpendingFundingSinkResult {
  const result = useQuery<Array<Partial<SpendingFunding>>>(
    `/bank_accounts/${ spending?.bankAccountId }/spending/${ spending?.spendingId }/funding`,
    {
      enabled: !!spending,
    }
  );
  return {
    ...result,
    result: (result?.data || []).map(item => new SpendingFunding(item)),
  };
}

export interface SpendingFundingCombined {
  funding: SpendingFunding;
  schedule: FundingSchedule;
}

export type SpendingFundingResult =
  {
    // Result includes all of the spending funding items for the specified spending object.
    result: Array<SpendingFundingCombined>;
    // Next is an array of all of the spending funding items that fall on the next funding day.
    // Next should always have at least one item. If it does not that is due to an error that
    // should be handled by the caller.
    next: Array<SpendingFundingCombined>;
  } & UseQueryResult<Array<Partial<SpendingFunding>>>;

export function useSpendingFunding(spending: Spending | null): SpendingFundingResult {
  const initial = useSpendingFundingSink(spending);
  const schedules = useFundingSchedulesSink();

  // Build the initial set of combined funding schedule information.
  const combined: Array<SpendingFundingCombined> = (initial.result || []).map(item => ({
    funding: item,
    schedule: schedules.result.get(item.fundingScheduleId),
  }));

  // Then group the funding by the next funding date.
  const next = combined
    .reduce((result: Record<number, Array<SpendingFundingCombined>>, item: SpendingFundingCombined) => {
      const date = item.schedule.nextOccurrence.unix();
      (result[date] = result[date] || []).push(item);
      return result;
    }, {});

  // Sort the funding dates in ascending order.
  const first = Object.keys(next).sort();

  return {
    ...initial,
    result: combined,
    // If we have at least one "next", then get the first group from the next.
    next: first.length > 0 ? next[first[0]] : [],
  };
}
