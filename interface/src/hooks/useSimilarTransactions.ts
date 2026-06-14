import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import type Transaction from '@monetr/interface/models/Transaction';
import TransactionCluster from '@monetr/interface/models/TransactionCluster';
import type { WithJsonValues } from '@monetr/interface/util/json';

export function useSimilarTransactions(transaction?: Transaction): UseQueryResult<TransactionCluster, unknown> {
  return useQuery<WithJsonValues<TransactionCluster>, unknown, TransactionCluster>({
    queryKey: [`/api/bank_accounts/${transaction?.bankAccountId}/transactions/${transaction?.transactionId}/similar`],
    enabled: Boolean(transaction),
    select: data => new TransactionCluster(data),
    retry: false,
  });
}
