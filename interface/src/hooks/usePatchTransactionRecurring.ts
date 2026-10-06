import { useMutation } from '@tanstack/react-query';

import type BankAccount from '@monetr/interface/models/BankAccount';
import type { ID } from '@monetr/interface/models/ID';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import type { WithJsonValues } from '@monetr/interface/util/json';
import type { Writable } from '@monetr/interface/util/readonly';
import request from '@monetr/interface/util/request';

// Only the links on a recurring transaction are writable, everything else about it is calculated by the server.
export type PatchTransactionRecurringRequest = Partial<Writable<TransactionRecurring>> & {
  transactionRecurringId: ID<TransactionRecurring>;
  bankAccountId: ID<BankAccount>;
};

export type PatchTransactionRecurringResponse = WithJsonValues<TransactionRecurring>;

export function usePatchTransactionRecurring(): (
  patch: PatchTransactionRecurringRequest,
) => Promise<PatchTransactionRecurringResponse> {
  const { mutateAsync } = useMutation({
    async mutationFn({
      transactionRecurringId,
      bankAccountId,
      ...recurring
    }: PatchTransactionRecurringRequest): Promise<PatchTransactionRecurringResponse> {
      return await request<PatchTransactionRecurringResponse>({
        method: 'PATCH',
        url: `/api/bank_accounts/${bankAccountId}/recurring/${transactionRecurringId}`,
        // undefined keys get dropped when this is stringified, a null still gets sent so the link can be cleared.
        data: recurring,
      }).then(result => result.data);
    },
    // The response already embeds whatever its linked to, so it can replace the cached recurring transaction as is.
    onSuccess: (result: PatchTransactionRecurringResponse, _input, _, { client: queryClient }) =>
      queryClient.setQueryData(
        [`/api/bank_accounts/${result.bankAccountId}/recurring/${result.transactionRecurringId}`],
        result,
      ),
  });

  return mutateAsync;
}
