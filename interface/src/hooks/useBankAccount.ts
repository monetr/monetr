import { useCallback, useMemo } from 'react';
import { type UseQueryResult, useQuery, useQueryClient } from '@tanstack/react-query';

import BankAccount from '@monetr/interface/models/BankAccount';

export function useBankAccount(bankAccountId?: string): UseQueryResult<BankAccount | undefined, unknown> {
  const queryClient = useQueryClient();
  const existingData = useMemo(() => queryClient.getQueryData<Array<BankAccount>>(['/bank_accounts']), [queryClient]);
  const initialData = useCallback(
    () =>
      // If the bank account is in our existing query state then use that.
      Array.isArray(existingData)
        ? existingData.find(item => item.bankAccountId === bankAccountId)
        : // Otherwise fall back to undefined.
          undefined,
    [existingData, bankAccountId],
  );
  const initialDataUpdatedAt = useCallback(
    () => queryClient.getQueryState(['/bank_accounts'])?.dataUpdatedAt,
    [queryClient],
  );

  return useQuery<Partial<BankAccount>, unknown, BankAccount | undefined>({
    queryKey: [`/bank_accounts/${bankAccountId}`],
    enabled: Boolean(bankAccountId), // Only request if we have a valid bank account ID to work with.
    select: data => Boolean(data) && new BankAccount(data),
    initialData,
    initialDataUpdatedAt,
  });
}
