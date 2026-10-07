import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import type { ID } from '@monetr/interface/models/ID';
import Transaction from '@monetr/interface/models/Transaction';
import type TransactionCluster from '@monetr/interface/models/TransactionCluster';
import type { WithJsonValues } from '@monetr/interface/util/json';

/**
 * useTransactionClusterTransactions returns the most recent 100 transactions in a similar transactions group. Unlike
 * useSimilarTransactions this is not paged, it is for when you want a quick look at the whole group at once.
 */
export function useTransactionClusterTransactions(
  transactionClusterId: ID<TransactionCluster> | null,
): UseQueryResult<Array<Transaction>, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Array<WithJsonValues<Transaction>>, unknown, Array<Transaction>>({
    queryKey: [
      'GET',
      `/api/bank_accounts/${selectedBankAccountId}/transactions`,
      { transaction_cluster_id: transactionClusterId, limit: 100 },
    ],
    enabled: Boolean(selectedBankAccountId && transactionClusterId),
    select: data => data.map(item => new Transaction(item)),
  });
}
