import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import Spending from '@monetr/interface/models/Spending';

export function useSpendings(): UseQueryResult<Array<Spending>, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Array<Partial<Spending>>, unknown, Array<Spending>>({
    queryKey: [`/bank_accounts/${selectedBankAccountId}/spending`],
    enabled: Boolean(selectedBankAccountId),
    initialData: [],
    initialDataUpdatedAt: 0,
    select: data => (data || []).map(item => new Spending(item)),
  });
}
