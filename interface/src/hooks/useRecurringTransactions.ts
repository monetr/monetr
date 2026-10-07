import { useCallback } from 'react';
import { type InfiniteData, type UseInfiniteQueryResult, useInfiniteQuery } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import type { WithJsonValues } from '@monetr/interface/util/json';

// Both of these are optional, leaving one off just means the server won't filter on it.
export interface RecurringTransactionsParams {
  direction?: TransactionRecurring['direction'];
  ended?: boolean;
}

export function useRecurringTransactions(
  params: RecurringTransactionsParams,
): UseInfiniteQueryResult<Array<TransactionRecurring>, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  const select = useCallback(
    (data: InfiniteData<Array<WithJsonValues<TransactionRecurring>>>) =>
      data.pages.flat().map(item => new TransactionRecurring(item)),
    [],
  );
  return useInfiniteQuery<Array<WithJsonValues<TransactionRecurring>>, unknown, Array<TransactionRecurring>>({
    queryKey: ['GET', `/api/bank_accounts/${selectedBankAccountId}/recurring`, params],
    initialPageParam: 0,
    getNextPageParam: (_, pages) => {
      // If there are no more pages then we should return null.
      if (pages.some(page => page.length < 25)) {
        return null;
      }
      // Otherwise we simply return the number of pages we have already requests times 25 since that is our page size.
      return pages.length * 25;
    },
    enabled: Boolean(selectedBankAccountId),
    // We want to flatten the data we return to the caller so that way it is easier to work with.
    select,
  });
}
