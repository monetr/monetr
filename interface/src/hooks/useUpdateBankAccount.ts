import { useCallback } from 'react';
import { useMutation } from '@tanstack/react-query';

import BankAccount from '@monetr/interface/models/BankAccount';
import type { ID } from '@monetr/interface/models/ID';
import type { WithJsonValues } from '@monetr/interface/util/json';
import type { Writable } from '@monetr/interface/util/readonly';
import request from '@monetr/interface/util/request';

export type PatchBankAccountRequest = Partial<Writable<Omit<BankAccount, 'accountType' | 'accountSubType'>>> & {
  bankAccountId: ID<BankAccount>;
};

export function useUpdateBankAccount(): (_bankAccount: PatchBankAccountRequest) => Promise<BankAccount> {
  const updateBankAccount = useCallback(
    async ({ bankAccountId, ...bankAccount }: PatchBankAccountRequest): Promise<BankAccount> => {
      return request<WithJsonValues<BankAccount>>({
        method: 'PATCH',
        url: `/api/bank_accounts/${bankAccountId}`,
        data: bankAccount,
      }).then(result => new BankAccount(result.data));
    },
    [],
  );

  const { mutateAsync } = useMutation({
    mutationFn: updateBankAccount,
    onSuccess: (updatedBankAccount: BankAccount, _var, _result, ctx) =>
      Promise.all([
        ctx.client.setQueryData(['/api/bank_accounts'], (previous: Array<WithJsonValues<BankAccount>>) =>
          previous.map(item => (item.bankAccountId === updatedBankAccount.bankAccountId ? updatedBankAccount : item)),
        ),
        ctx.client.setQueryData([`/api/bank_accounts/${updatedBankAccount.bankAccountId}`], updatedBankAccount),
        ctx.client.invalidateQueries({ queryKey: [`/api/bank_accounts/${updatedBankAccount.bankAccountId}/balances`] }),
      ]),
  });

  return mutateAsync;
}
