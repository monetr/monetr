import { useQuery, UseQueryResult } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import FundingSchedule from '@monetr/interface/models/FundingSchedule';

export function useFundingSchedules(): UseQueryResult<Array<FundingSchedule>, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Array<Partial<FundingSchedule>>, unknown, Array<FundingSchedule>>({
    queryKey: [`/bank_accounts/${selectedBankAccountId}/funding_schedules`],
    enabled: Boolean(selectedBankAccountId),
    select: data => data.map(item => new FundingSchedule(item)),
  });
}
