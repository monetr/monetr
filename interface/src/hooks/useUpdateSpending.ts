import { useMutation, useQueryClient } from '@tanstack/react-query';

import Spending from '@monetr/interface/models/Spending';
import request from '@monetr/interface/util/request';

export function useUpdateSpending(): (_spending: Spending) => Promise<void> {
  const queryClient = useQueryClient();

  async function updateSpending(spending: Spending): Promise<Spending> {
    return request()
      .put<Partial<Spending>>(`/bank_accounts/${spending.bankAccountId}/spending/${spending.spendingId}`, spending)
      .then(result => new Spending(result?.data));
  }

  const { mutate } = useMutation({
    mutationFn: updateSpending,
    onSuccess: (updatedSpending: Spending) =>
      Promise.all([
        queryClient.setQueryData(
          [`/bank_accounts/${updatedSpending.bankAccountId}/spending`],
          (previous: Array<Partial<Spending>>) =>
            previous.map(item => (item.spendingId === updatedSpending.spendingId ? updatedSpending : item)),
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

  return async (spending: Spending): Promise<void> => {
    return mutate(spending);
  };
}
