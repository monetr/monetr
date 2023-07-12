import { useQuery, UseQueryResult } from 'react-query';
import shallow from 'zustand/shallow';

import { useLinks } from 'hooks/links';
import useStore from 'hooks/store';
import BankAccount from 'models/BankAccount';
import { useLocation, useParams } from 'react-router-dom';

export type BankAccountsResult =
  { result: Map<number, BankAccount> }
  & UseQueryResult<Array<Partial<BankAccount>>>;

export function useBankAccountsSink(): BankAccountsResult {
  const links = useLinks();
  const result = useQuery<Array<Partial<BankAccount>>>('/bank_accounts', {
    enabled: !!links && links.size > 0,
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

export type CurrentBankAccountResult =
  { result: BankAccount | null }
  & UseQueryResult<Partial<BankAccount>>;

export function useSelectedBankAccount(): CurrentBankAccountResult {
  const location = useLocation();
  const { bankAccountId: id } = useParams();
  const bankAccountId = +id || null;
  // If we do not have a valid numeric bank account ID, but an ID was specified then something is wrong.
  // if (!bankAccountId && id) {
  //   throw Error(`invalid bank account ID specified: "${id}" is not a valid bank account ID`);
  // }

  const result = useQuery<Partial<BankAccount>>(
    `/bank_accounts/${ bankAccountId }`,
    {
      enabled: !!bankAccountId && !!location, // Only request if we have a valid numeric bank account ID to work with.
    }
  );

  const thing = {
    ...result,
    result: !!result.data ? new BankAccount(result.data) : null,
  };

  console.log('producer', thing, !!bankAccountId);
  return thing;
}

export function useSelectedBankAccountId(): number | null {
  const { result: bankAccount } = useSelectedBankAccount();

  return bankAccount?.bankAccountId || null;
}

