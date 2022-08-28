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
    enabled: links.size > 0,
  });
  return {
    ...result,
    result: new Map(result?.data?.map(item => {
      const bankAccount = new BankAccount(item);
      return [bankAccount.bankAccountId, bankAccount];
    })),
  };
}

export function useBankAccounts(): Map<number, BankAccount> {
  const { result: bankAccounts } = useBankAccountsSink();
  return bankAccounts;
}

export function useSelectedBankAccountId(): number | null {
  const { selectedBankAccountId, setCurrentBankAccount } = useStore(state => ({
    selectedBankAccountId: state.selectedBankAccountId,
    setCurrentBankAccount: state.setCurrentBankAccount,
  }), shallow);
  const { isLoading, result: bankAccounts } = useBankAccountsSink();

  if (isLoading) {
    return selectedBankAccountId;
  }

  if (!isLoading && !bankAccounts.has(selectedBankAccountId)) {
    if (bankAccounts.size === 0) {
      return null;
    }

    const id = Array.from(bankAccounts.keys())[0];
    setCurrentBankAccount(id);
    return id;
  }

  return selectedBankAccountId;
}
