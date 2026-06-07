import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import Spending from '@monetr/interface/models/Spending';
import type { WithJsonValues } from '@monetr/interface/util/json';

export function useSpendings(): UseQueryResult<Array<Spending>, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Array<WithJsonValues<Spending>>, unknown, Array<Spending>>({
    queryKey: [`/api/bank_accounts/${selectedBankAccountId}/spending`],
    enabled: Boolean(selectedBankAccountId),
    select: data => (data || []).map(item => new Spending(item)),
  });
}
