import { useQuery, UseQueryResult } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';

export function useNextFundingForecast(fundingScheduleId: string): UseQueryResult<number, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Partial<{ nextContribution: number }>, unknown, number>({
    queryKey: [
      `/bank_accounts/${ selectedBankAccountId }/forecast/next_funding`,
      { fundingScheduleId },
    ],
    enabled: Boolean(selectedBankAccountId),
    select: data => data.nextContribution,
  });
}

