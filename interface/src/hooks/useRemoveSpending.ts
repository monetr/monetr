import { useMutation, useQueryClient } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import Spending from '@monetr/interface/models/Spending';
import request from '@monetr/interface/util/request';

export function useRemoveSpending(): (_spendingId: string) => Promise<unknown> {
  const queryClient = useQueryClient();
  const selectedBankAccountId = useSelectedBankAccountId();

  async function removeSpending(spendingId: string): Promise<string> {
    return request()
      .delete(`/bank_accounts/${selectedBankAccountId}/spending/${spendingId}`)
      .then(() => spendingId);
  }

  const mutation = useMutation({
    mutationFn: removeSpending,
    onSuccess: (removedSpendingId: string) =>
      Promise.all([
        queryClient.setQueryData(
          [`/bank_accounts/${selectedBankAccountId}/spending`],
          (previous: Array<Partial<Spending>>) => previous.filter(item => item.spendingId !== removedSpendingId),
        ),
        queryClient.removeQueries({
          queryKey: [`/bank_accounts/${selectedBankAccountId}/spending/${removedSpendingId}`],
        }),
        queryClient.invalidateQueries({ queryKey: [`/bank_accounts/${selectedBankAccountId}/balances`] }),
        queryClient.invalidateQueries({ queryKey: [`/bank_accounts/${selectedBankAccountId}/forecast`] }),
        queryClient.invalidateQueries({ queryKey: [`/bank_accounts/${selectedBankAccountId}/forecast/next_funding`] }),
        queryClient.invalidateQueries({ queryKey: [`/bank_accounts/${selectedBankAccountId}/transactions`] }),
      ]),
  });

  return mutation.mutateAsync;
}
