import { useMutation, useQueryClient } from '@tanstack/react-query';

import BankAccount from '@monetr/interface/models/BankAccount';
import request from '@monetr/interface/util/request';

export function useArchiveBankAccount(): (_bankAccountId: string) => Promise<string> {
  const queryClient = useQueryClient();

  async function archiveBankAccount(bankAccountId: string): Promise<string> {
    return request()
      .delete<Partial<BankAccount>>(`/bank_accounts/${bankAccountId}`)
      .then(() => bankAccountId);
  }

  const mutate = useMutation({
    mutationFn: archiveBankAccount,
    onSuccess: (bankAccountId: string) =>
      Promise.all([
        queryClient.setQueryData(['/bank_accounts'], (previous: Array<Partial<BankAccount>>) =>
          previous.filter(item => item.bankAccountId !== bankAccountId),
        ),
        queryClient.removeQueries({ queryKey: [`/bank_accounts/${bankAccountId}`] }),
      ]),
  });

  return mutate.mutateAsync;
}
