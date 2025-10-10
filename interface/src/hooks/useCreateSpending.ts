import { useMutation, useQueryClient } from '@tanstack/react-query';

import Spending from '@monetr/interface/models/Spending';
import request from '@monetr/interface/util/request';

export function useCreateSpending(): (_spending: Spending) => Promise<Spending> {
  const queryClient = useQueryClient();

  async function createSpending(spending: Spending): Promise<Spending> {
    return request()
      .post<Partial<Spending>>(`/bank_accounts/${ spending.bankAccountId }/spending`, spending)
      .then(result => new Spending(result?.data));
  }

  const mutation = useMutation({
    mutationFn: createSpending,
    onSuccess: (created: Spending) => Promise.all([
      queryClient.setQueryData(
        [`/bank_accounts/${ created.bankAccountId }/spending`],
        (previous: Array<Partial<Spending>>) => (previous || []).concat(created),
      ),
      queryClient.setQueryData(
        [`/bank_accounts/${ created.bankAccountId}/spending/${ created.spendingId}`],
        created,
      ),
      queryClient.invalidateQueries({ queryKey: [`/bank_accounts/${ created.bankAccountId }/balances`] }),
      queryClient.invalidateQueries({ queryKey: [`/bank_accounts/${ created.bankAccountId }/forecast`] }),
      queryClient.invalidateQueries({ queryKey: [`/bank_accounts/${ created.bankAccountId }/forecast/next_funding`] }),
    ]),
  });

  return mutation.mutateAsync;
}
