import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import type { ID } from '@monetr/interface/models/ID';
import TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import type { WithJsonValues } from '@monetr/interface/util/json';

export function useRecurringTransaction(
  transactionRecurringId: ID<TransactionRecurring> | null,
): UseQueryResult<TransactionRecurring, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<WithJsonValues<TransactionRecurring>, unknown, TransactionRecurring>({
    queryKey: [`/api/bank_accounts/${selectedBankAccountId}/recurring/${transactionRecurringId}`],
    enabled: Boolean(transactionRecurringId),
    select: data => new TransactionRecurring(data),
  });
}
