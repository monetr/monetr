import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import type { ID } from '@monetr/interface/models/ID';
import TransactionCluster from '@monetr/interface/models/TransactionCluster';
import type { WithJsonValues } from '@monetr/interface/util/json';

export function useTransactionCluster(
  transactionClusterId: ID<TransactionCluster> | null,
): UseQueryResult<TransactionCluster, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<WithJsonValues<TransactionCluster>, unknown, TransactionCluster>({
    queryKey: [`/api/bank_accounts/${selectedBankAccountId}/similar/${transactionClusterId}`],
    enabled: Boolean(transactionClusterId),
    select: data => new TransactionCluster(data),
  });
}
