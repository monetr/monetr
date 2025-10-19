import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import FundingSchedule from '@monetr/interface/models/FundingSchedule';

export function useFundingSchedule(fundingScheduleId?: string): UseQueryResult<FundingSchedule | null, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Partial<FundingSchedule>, unknown, FundingSchedule | null>({
    queryKey: [`/bank_accounts/${selectedBankAccountId}/funding_schedules/${fundingScheduleId}`],
    enabled: Boolean(selectedBankAccountId) && Boolean(fundingScheduleId),
    select: data => (data?.fundingScheduleId ? new FundingSchedule(data) : null),
  });
}
