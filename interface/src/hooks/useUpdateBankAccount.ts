import { useMutation, useQueryClient } from '@tanstack/react-query';

import BankAccount from '@monetr/interface/models/BankAccount';
import type { WithJsonValues } from '@monetr/interface/util/json';
import request from '@monetr/interface/util/request';

export interface UpdateBankAccountRequest {
  bankAccountId: string;
  name: string;
  currency: string;
  availableBalance?: number;
  currentBalance?: number;
  limitBalance?: number;
}

export function useUpdateBankAccount(): (_bankAccount: UpdateBankAccountRequest) => Promise<BankAccount> {
  const queryClient = useQueryClient();

  async function updateBankAccount({ bankAccountId, ...updates }: UpdateBankAccountRequest): Promise<BankAccount> {
    return request<WithJsonValues<BankAccount>>({
      method: 'PUT',
      url: `/api/bank_accounts/${bankAccountId}`,
      data: updates,
    }).then(result => new BankAccount(result?.data));
  }

  const mutate = useMutation({
    mutationFn: updateBankAccount,
    onSuccess: (updatedBankAccount: BankAccount) =>
      Promise.all([
        queryClient.setQueryData(['/api/bank_accounts'], (previous: Array<WithJsonValues<BankAccount>>) =>
          previous.map(item => (item.bankAccountId === updatedBankAccount.bankAccountId ? updatedBankAccount : item)),
        ),
        queryClient.setQueryData([`/api/bank_accounts/${updatedBankAccount.bankAccountId}`], updatedBankAccount),
      ]),
  });

  return mutate.mutateAsync;
}
