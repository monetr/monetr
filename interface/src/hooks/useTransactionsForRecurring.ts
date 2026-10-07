import { useCallback } from 'react';
import { type InfiniteData, type UseInfiniteQueryResult, useInfiniteQuery } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import type { ID } from '@monetr/interface/models/ID';
import Transaction from '@monetr/interface/models/Transaction';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import type { WithJsonValues } from '@monetr/interface/util/json';

// useTransactionsForRecurring pages through the transactions monetr matched to a recurring transaction, newest first.
// Same idea as useSimilarTransactions but for just the ones on the schedule. The limit is the page size, so something
// that only shows a few can ask for just those.
export function useTransactionsForRecurring(
  transactionRecurringId: ID<TransactionRecurring> | undefined,
  limit: number,
): UseInfiniteQueryResult<Array<Transaction>, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  const select = useCallback(
    (data: InfiniteData<Array<WithJsonValues<Transaction>>>) => data.pages.flat().map(item => new Transaction(item)),
    [],
  );
  return useInfiniteQuery<Array<WithJsonValues<Transaction>>, unknown, Array<Transaction>>({
    queryKey: [
      'GET',
      `/api/bank_accounts/${selectedBankAccountId}/transactions`,
      {
        transaction_recurring_id: transactionRecurringId,
        limit: limit,
      },
    ],
    initialPageParam: 0,
    getNextPageParam: (_, pages) => {
      // If there are no more pages then we should return null.
      if (pages.some(page => page.length < limit)) {
        return null;
      }
      // Otherwise we simply return the number of pages we have already requests times the page size.
      return pages.length * limit;
    },
    enabled: Boolean(selectedBankAccountId && transactionRecurringId),
    // We want to flatten the data we return to the caller so that way it is easier to work with.
    select,
  });
}
