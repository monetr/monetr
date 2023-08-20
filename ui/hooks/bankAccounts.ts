import { useMatch } from 'react-router-dom';
import { useQuery, UseQueryResult } from '@tanstack/react-query';

import { useLinksSink } from 'hooks/links';
import BankAccount from 'models/BankAccount';

export type BankAccountsResult =
  { result: Map<number, BankAccount> }
  & UseQueryResult<Array<Partial<BankAccount>>>;

export function useBankAccountsSink(): BankAccountsResult {
  const { data: links } = useLinksSink();
  const result = useQuery<Array<Partial<BankAccount>>>(['/bank_accounts'], {
    enabled: !!links && links.length > 0,
  });
  return {
    ...result,
    result: new Map((result?.data || []).map(item => {
      const bankAccount = new BankAccount(item);
      return [bankAccount.bankAccountId, bankAccount];
    })),
  };
}

export function useBankAccounts(): Map<number, BankAccount> {
  const { result: bankAccounts } = useBankAccountsSink();
  return bankAccounts;
}

export interface SelectedBankAccountResult {
  isLoading: boolean;
  isError: boolean;
  bankAccount: BankAccount | null;
}

export function useSelectedBankAccount(): UseQueryResult<BankAccount | undefined> {
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
    }
  );
}

export function useSelectedBankAccountId(): number | undefined {
  const { data: bankAccount } = useSelectedBankAccount();
  return bankAccount?.bankAccountId;
}

