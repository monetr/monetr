import { useMutation } from '@tanstack/react-query';

import Spending from '@monetr/interface/models/Spending';
import request from '@monetr/interface/util/request';

export function useUpdateSpending(): (_spending: Spending) => Promise<Spending> {
  const { mutateAsync } = useMutation({
    mutationFn: async (spending: Spending): Promise<Spending> =>
      request()
        .put<Partial<Spending>>(`/bank_accounts/${spending.bankAccountId}/spending/${spending.spendingId}`, spending)
        .then(result => new Spending(result?.data)),
    onSuccess: (updatedSpending: Spending, _variables, _onMutateResult, { client: queryClient }) =>
      Promise.all([
        queryClient.setQueryData(
          [`/bank_accounts/${updatedSpending.bankAccountId}/spending`],
          (previous: Array<Partial<Spending>>) =>
            (previous ?? []).map(item => (item.spendingId === updatedSpending.spendingId ? updatedSpending : item)),
        ),
        queryClient.setQueryData(
          [`/bank_accounts/${updatedSpending.bankAccountId}/spending/${updatedSpending.spendingId}`],
          updatedSpending,
        ),
        // TODO Under what circumstances do we need to invalidate balances for a spending update?
        queryClient.invalidateQueries({
          queryKey: [`/bank_accounts/${updatedSpending.bankAccountId}/balances`],
        }),
        queryClient.invalidateQueries({
          queryKey: [`/bank_accounts/${updatedSpending.bankAccountId}/forecast`],
        }),
        queryClient.invalidateQueries({
          queryKey: [`/bank_accounts/${updatedSpending.bankAccountId}/forecast/next_funding`],
        }),
      ]),
  });

  return mutateAsync;
}
