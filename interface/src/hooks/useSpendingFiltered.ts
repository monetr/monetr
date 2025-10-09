import { useQuery, UseQueryResult } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import Spending, { SpendingType } from '@monetr/interface/models/Spending';

export function useSpendingFiltered(kind: SpendingType): UseQueryResult<Array<Spending>, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Array<Partial<Spending>>, unknown, Array<Spending>>({
    // Use the same query key so that way the request is not sent again if the data is already in the cache.
    queryKey: [`/bank_accounts/${selectedBankAccountId}/spending`],
    enabled: Boolean(selectedBankAccountId),
    initialData: [],
    initialDataUpdatedAt: 0,
    select: data => (data || [])
      .map(item => new Spending(item))
      // Filter the data by the kind specified!
      .filter(item => item.spendingType === kind),
  });
}
