import { useMatch } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient, UseQueryResult } from '@tanstack/react-query';

import { useLinks } from '@monetr/interface/hooks/links';
import BankAccount from '@monetr/interface/models/BankAccount';
import request from '@monetr/interface/util/request';

export function useBankAccounts(): UseQueryResult<Array<BankAccount>> {
  const { data: links } = useLinks();
  return useQuery<Array<Partial<BankAccount>>, unknown, Array<BankAccount>>(['/bank_accounts'], {
    enabled: !!links && links.length > 0,
    select: data => data.map(item => new BankAccount(item)),
  });
}

export interface CreateBankAccountRequest {
  linkId: number;
  name: string;
  mask?: string;
  availableBalance: number;
  currentBalance: number;
  accountType: string;
  accountSubType: string;
}

export function useCreateBankAccount(): (_bankAccount: CreateBankAccountRequest) => Promise<BankAccount> {
  const queryClient = useQueryClient();

  async function createBankAccount(newBankAccount: CreateBankAccountRequest): Promise<BankAccount> {
    return request()
      .post<Partial<BankAccount>>('/bank_accounts', newBankAccount)
      .then(result => new BankAccount(result?.data));
  }

  const mutate = useMutation(
    createBankAccount,
    {
      onSuccess: (newBankAccount: BankAccount) => Promise.all([
        queryClient.setQueriesData(
          ['/bank_accounts'],
          (previous: Array<Partial<BankAccount>>) => (previous ?? []).concat(newBankAccount),
        ),
        queryClient.setQueriesData(
          [`/bank_accounts/${newBankAccount.bankAccountId}`],
          newBankAccount,
        ),
      ]),
    }
  );

  return mutate.mutateAsync;
}

export function useSelectedBankAccount(): UseQueryResult<BankAccount | undefined> {
  const queryClient = useQueryClient();
  const match = useMatch('/bank/:bankId/*');
  const bankAccountId = +match?.params?.bankId || null;

  // If we do not have a valid numeric bank account ID, but an ID was specified then something is wrong.
  if (!bankAccountId && match?.params?.bankId) {
    throw Error(`invalid bank account ID specified: "${match?.params?.bankId}" is not a valid bank account ID`);
  }

  return useQuery<Partial<BankAccount>, unknown, BankAccount | undefined>(
    [`/bank_accounts/${bankAccountId}`],
    {
      enabled: !!bankAccountId, // Only request if we have a valid numeric bank account ID to work with.
      select: data => !!data && new BankAccount(data),
      initialData: () => queryClient
        .getQueryData<Array<BankAccount>>(['/bank_accounts'])
        ?.find(item => item.bankAccountId === bankAccountId),
      initialDataUpdatedAt: () => queryClient
        .getQueryState(['/bank_accounts'])?.dataUpdatedAt,
    }
  );
}

export function useSelectedBankAccountId(): number | undefined {
  const { data: bankAccount } = useSelectedBankAccount();
  return bankAccount?.bankAccountId;
}

