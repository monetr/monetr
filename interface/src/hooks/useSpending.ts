import { useQuery, useQueryClient, UseQueryResult } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import Spending from '@monetr/interface/models/Spending';

/**
 * useSpending is used to retrieve data on a single spending object. It will however, hydrate it's state using data from
 * all of the spending objects that have already been queried via a list endpoint. If the desired spending object is not
 * in that list though, then it will make a request to retrieve that spending object's details.
 */
export function useSpending(spendingId?: string): UseQueryResult<Spending, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  const queryClient = useQueryClient();
  const existingData = queryClient.getQueryData<Array<Spending>>(
    [`/bank_accounts/${selectedBankAccountId}/spending`],
  );

  return useQuery<Partial<Spending>, unknown, Spending>({
    queryKey: [`/bank_accounts/${selectedBankAccountId}/spending/${spendingId}`],
    enabled: Boolean(selectedBankAccountId) && Boolean(spendingId),
    initialData: () => Array.isArray(existingData) ?
      // If the spending is in our existing query state then use that.
      existingData.find(item => item.spendingId === spendingId) :
      // Otherwise fall back to undefined.
      undefined,
    initialDataUpdatedAt: () => queryClient.getQueryState(
      [`/bank_accounts/${selectedBankAccountId}/spending`],
    )?.dataUpdatedAt,
    select: data => new Spending(data),
  });
}
