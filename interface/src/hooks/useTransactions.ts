import { useInfiniteQuery, type UseInfiniteQueryResult } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import Transaction from '@monetr/interface/models/Transaction';

export function useTransactions(): UseInfiniteQueryResult<Array<Transaction>, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useInfiniteQuery<Array<Partial<Transaction>>, unknown, Array<Transaction>>({
    queryKey: [`/bank_accounts/${selectedBankAccountId}/transactions`],
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
    select: data => data.pages.flatMap(x => x).map(item => new Transaction(item)),
  });
}
