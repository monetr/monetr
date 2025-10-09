import { useMatch } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient, UseQueryResult } from '@tanstack/react-query';

import BankAccount from '@monetr/interface/models/BankAccount';
import request from '@monetr/interface/util/request';

export interface UpdateBankAccountRequest {
  bankAccountId: string;
  name: string;
  currency: string;
}

export function useUpdateBankAccount(): (_bankAccount: UpdateBankAccountRequest) => Promise<BankAccount> {
  const queryClient = useQueryClient();

  async function updateBankAccount({ bankAccountId, ...updates }: UpdateBankAccountRequest): Promise<BankAccount> {
    return request()
      .put<Partial<BankAccount>>(`/bank_accounts/${bankAccountId}`, updates)
      .then(result => new BankAccount(result?.data));
  }

  const mutate = useMutation(
    updateBankAccount,
    {
      onSuccess: (updatedBankAccount: BankAccount) => Promise.all([
        queryClient.setQueriesData(
          ['/bank_accounts'],
          (previous: Array<Partial<BankAccount>>) =>
            previous.map(item => item.bankAccountId === updatedBankAccount.bankAccountId ? updatedBankAccount : item),
        ),
        queryClient.setQueriesData(
          [`/bank_accounts/${updatedBankAccount.bankAccountId}`],
          updatedBankAccount,
        ),
      ]),
    }
  );

  return mutate.mutateAsync;
}

export function useArchiveBankAccount(): (_bankAccountId: string) => Promise<string> {
  const queryClient = useQueryClient();

  async function archiveBankAccount(bankAccountId: string): Promise<string> {
    return request()
      .delete<Partial<BankAccount>>(`/bank_accounts/${bankAccountId}`)
      .then(() => bankAccountId);
  }

  const mutate = useMutation(
    archiveBankAccount,
    {
      onSuccess: (bankAccountId: string) => Promise.all([
        queryClient.setQueryData(
          ['/bank_accounts'],
          (previous: Array<Partial<BankAccount>>) => previous.filter(item => item.bankAccountId !== bankAccountId),
        ),
        queryClient.removeQueries([`/bank_accounts/${bankAccountId}`]),
      ]),
    }
  );

  return mutate.mutateAsync;
}

export function useSelectedBankAccount(): UseQueryResult<BankAccount | undefined> {
  const queryClient = useQueryClient();
  const match = useMatch('/bank/:bankId/*');
  const bankAccountId = match?.params?.bankId || null;

  // If we do not have a valid numeric bank account ID, but an ID was specified then something is wrong.
  if (!bankAccountId && match?.params?.bankId) {
    throw Error(`invalid bank account ID specified: "${match?.params?.bankId}" is not a valid bank account ID`);
  }

  const existingData = queryClient.getQueryData<Array<BankAccount>>(['/bank_accounts']);

  return useQuery<Partial<BankAccount>, unknown, BankAccount | undefined>(
    [`/bank_accounts/${bankAccountId}`],
    {
      enabled: !!bankAccountId, // Only request if we have a valid numeric bank account ID to work with.
      select: data => !!data && new BankAccount(data),
      // If the bank account is in our existing query state then use that.
      initialData: () => Array.isArray(existingData) ?
        existingData.find(item => item.bankAccountId === bankAccountId) :
        // Otherwise fall back to undefined.
        undefined,
      initialDataUpdatedAt: () => queryClient.getQueryState(['/bank_accounts'])?.dataUpdatedAt,
    }
  );
}

export function useSelectedBankAccountId(): string | undefined {
  const { data: bankAccount } = useSelectedBankAccount();
  return bankAccount?.bankAccountId;
}

