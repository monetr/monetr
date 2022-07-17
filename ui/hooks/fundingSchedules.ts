import { useQuery, UseQueryResult } from 'react-query';

import { useBankAccountsSink } from 'hooks/bankAccounts';
import FundingSchedule from 'models/FundingSchedule';

export type FundingSchedulesResult =
  { result: Map<number, FundingSchedule> }
  & UseQueryResult<Array<Partial<FundingSchedule>>>;

export function useFundingSchedulesSink(): FundingSchedulesResult {
  const { result: { selectedBankAccountId } } = useBankAccountsSink();
  const result = useQuery<Array<Partial<FundingSchedule>>>(
    `/bank_accounts/${ selectedBankAccountId }/funding_schedules`,
    {
      enabled: !!selectedBankAccountId,
    },
  );
  return {
    ...result,
    result: new Map(result?.data?.map(item => {
      const fundingSchedule = new FundingSchedule(item);
      return [fundingSchedule.fundingScheduleId, fundingSchedule];
    })),
  };
}

export function useFundingSchedule(fundingScheduleId: number): FundingSchedule | null {
  const { result } = useFundingSchedulesSink();
  return result.get(fundingScheduleId) || null;
}
