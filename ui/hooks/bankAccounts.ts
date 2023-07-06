import { useQuery, UseQueryResult } from 'react-query';
import shallow from 'zustand/shallow';

import { useLinks } from 'hooks/links';
import useStore from 'hooks/store';
import BankAccount from 'models/BankAccount';

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

export function useSelectedBankAccount(): SelectedBankAccountResult {
  const { selectedBankAccountId, setCurrentBankAccount } = useStore(state => ({
    selectedBankAccountId: state.selectedBankAccountId,
    setCurrentBankAccount: state.setCurrentBankAccount,
  }), shallow);
  const { isError, isLoading, result: bankAccounts } = useBankAccountsSink();
  if (isLoading || isError) {
    return {
      isLoading,
      isError,
      bankAccount: null,
    };
  }

  if (!bankAccounts.has(selectedBankAccountId)) {
    if (bankAccounts.size === 0) {
      return {
        isLoading: false,
        isError: true,
        bankAccount: null,
      };
    }

    const id = Array.from(bankAccounts.keys())[0];
    setCurrentBankAccount(id);
    return {
      isLoading: false,
      isError: false,
      bankAccount: bankAccounts.get(id),
    };
  }

  return {
    isLoading: false,
    isError: false,
    bankAccount: bankAccounts.get(selectedBankAccountId),
  };
}

export function useSelectedBankAccountId(): number | null {
  const { bankAccount } = useSelectedBankAccount();

  return bankAccount?.bankAccountId || null;
}

