import { useQuery, UseQueryResult } from '@tanstack/react-query';

import Transaction from '@monetr/interface/models/Transaction';
import TransactionCluster from '@monetr/interface/models/TransactionCluster';

export function useSimilarTransactions(transaction?: Transaction): UseQueryResult<TransactionCluster, unknown> {
  return useQuery<Partial<TransactionCluster>, unknown, TransactionCluster>({
    queryKey: [`/bank_accounts/${transaction?.bankAccountId}/transactions/${transaction?.transactionId}/similar`],
    enabled: Boolean(transaction),
    select: data => new TransactionCluster(data),
    retry: false,
  });
}

