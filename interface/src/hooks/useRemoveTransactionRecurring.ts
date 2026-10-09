import { useMutation } from '@tanstack/react-query';

import type BankAccount from '@monetr/interface/models/BankAccount';
import type { ID } from '@monetr/interface/models/ID';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import request from '@monetr/interface/util/request';

export interface RemoveTransactionRecurringRequest {
  transactionRecurringId: ID<TransactionRecurring>;
  bankAccountId: ID<BankAccount>;
}

export function useRemoveTransactionRecurring(): (remove: RemoveTransactionRecurringRequest) => Promise<void> {
  const { mutateAsync } = useMutation({
    async mutationFn({ transactionRecurringId, bankAccountId }: RemoveTransactionRecurringRequest): Promise<void> {
      await request({
        method: 'DELETE',
        url: `/api/bank_accounts/${bankAccountId}/recurring/${transactionRecurringId}`,
      });
    },
    onSuccess: (_, { transactionRecurringId, bankAccountId }, __, { client: queryClient }) =>
      Promise.all([
        // Nothing comes back from the delete, so refetch the recurring transaction to pick up that it's deleted now.
        queryClient.invalidateQueries({
          queryKey: ['GET', `/api/bank_accounts/${bankAccountId}/recurring/${transactionRecurringId}`],
        }),
        queryClient.invalidateQueries({
          queryKey: ['GET', `/api/bank_accounts/${bankAccountId}/recurring`],
        }),
      ]),
  });

  return mutateAsync;
}
