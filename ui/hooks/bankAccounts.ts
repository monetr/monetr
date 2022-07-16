import { useQuery, UseQueryResult } from 'react-query';
import shallow from 'zustand/shallow';

import { useLinks } from 'hooks/links';
import useStore from 'hooks/store';
import BankAccount from 'models/BankAccount';

export type BankAccountsResult =
  {
    result: {
      setCurrentBankAccount: (_bankAccountId: number) => void;
      selectedBankAccountId: number | null;
      bankAccounts: Map<number, BankAccount>;
    }
  }
  & UseQueryResult<Array<Partial<BankAccount>>>;

export function useBankAccountsSink(): BankAccountsResult {
  const links = useLinks();
  const result = useQuery<Array<Partial<BankAccount>>>('/api/bank_accounts', {
    enabled: links.size > 0,
  });
  const { selectedBankAccountId, setCurrentBankAccount } = useStore(state => ({
    selectedBankAccountId: state.selectedBankAccountId,
    setCurrentBankAccount: state.setCurrentBankAccount,
  }), shallow);
  return {
    ...result,
    result: {
      setCurrentBankAccount,
      selectedBankAccountId,
      bankAccounts: new Map(result?.data?.map(item => {
        const bankAccount = new BankAccount(item);
        return [bankAccount.bankAccountId, bankAccount];
      })),
    },
  };
}

export function useBankAccounts(): Map<number, BankAccount> {
  const { result: { bankAccounts } } = useBankAccountsSink();
  return bankAccounts;
}

