import { useMutation, useQueryClient } from '@tanstack/react-query';

import BankAccount, { type BankAccountSubType, type BankAccountType } from '@monetr/interface/models/BankAccount';
import request from '@monetr/interface/util/request';

export interface CreateBankAccountRequest {
  linkId: string;
  lunchFlowBankAccountId?: string;
  name: string;
  mask?: string;
  availableBalance: number;
  currentBalance: number;
  accountType: BankAccountType;
  accountSubType: BankAccountSubType;
  currency: string;
}

export function useCreateBankAccount(): (_bankAccount: CreateBankAccountRequest) => Promise<BankAccount> {
  const queryClient = useQueryClient();

  async function createBankAccount(newBankAccount: CreateBankAccountRequest): Promise<BankAccount> {
    return request()
      .post<Partial<BankAccount>>('/bank_accounts', newBankAccount)
      .then(result => new BankAccount(result?.data));
  }

  const mutate = useMutation({
    mutationFn: createBankAccount,
    onSuccess: (newBankAccount: BankAccount) =>
      Promise.all([
        queryClient.setQueryData(['/bank_accounts'], (previous: Array<Partial<BankAccount>>) =>
          (previous ?? []).concat(newBankAccount),
        ),
        queryClient.setQueryData([`/bank_accounts/${newBankAccount.bankAccountId}`], newBankAccount),
      ]),
  });

  return mutate.mutateAsync;
}
