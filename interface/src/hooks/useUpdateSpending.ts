import { useMutation } from '@tanstack/react-query';

import type BankAccount from '@monetr/interface/models/BankAccount';
import type { ID } from '@monetr/interface/models/ID';
import Spending from '@monetr/interface/models/Spending';
import type { WithJsonValues } from '@monetr/interface/util/json';
import type { Writable } from '@monetr/interface/util/readonly';
import request from '@monetr/interface/util/request';

export type PatchSpendingRequest = Partial<Writable<Spending>> & {
  spendingId: ID<Spending>;
  bankAccountId: ID<BankAccount>;
};

export function useUpdateSpending(): (_spending: PatchSpendingRequest) => Promise<Spending> {
  const { mutateAsync } = useMutation({
    async mutationFn({ spendingId, bankAccountId, ...spending }: PatchSpendingRequest): Promise<Spending> {
      return await request<WithJsonValues<Spending>>({
        method: 'PATCH',
        url: `/api/bank_accounts/${bankAccountId}/spending/${spendingId}`,
        data: spending,
      }).then(result => new Spending(result.data));
    },
    onSuccess: (updatedSpending: Spending, _variables, _onMutateResult, { client: queryClient }) =>
      Promise.all([
        queryClient.setQueryData(
          [`/api/bank_accounts/${updatedSpending.bankAccountId}/spending`],
          (previous: Array<Partial<Spending>>) =>
            (previous ?? []).map(item => (item.spendingId === updatedSpending.spendingId ? updatedSpending : item)),
        ),
        queryClient.setQueryData(
          [`/api/bank_accounts/${updatedSpending.bankAccountId}/spending/${updatedSpending.spendingId}`],
          updatedSpending,
        ),
        // TODO Under what circumstances do we need to invalidate balances for a spending update?
        queryClient.invalidateQueries({
          queryKey: [`/api/bank_accounts/${updatedSpending.bankAccountId}/balances`],
        }),
        queryClient.invalidateQueries({
          queryKey: [`/api/bank_accounts/${updatedSpending.bankAccountId}/forecast`],
        }),
        queryClient.invalidateQueries({
          queryKey: [`/api/bank_accounts/${updatedSpending.bankAccountId}/forecast/next_funding`],
        }),
      ]),
  });

  return mutateAsync;
}
