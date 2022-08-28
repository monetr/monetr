import { useQuery } from 'react-query';

import { useSelectedBankAccountId } from 'hooks/bankAccounts';
import Balance from 'models/Balance';

export function useBalance(bankAccountId: number): Balance | null {
  const result = useQuery<Partial<Balance>>(`/bank_accounts/${ bankAccountId }/balances`);
  return result?.data && new Balance(result?.data);
}

export function useCurrentBalance(): Balance | null {
  const selectedBankAccountId = useSelectedBankAccountId();
  const result = useQuery<Partial<Balance>>(`/bank_accounts/${ selectedBankAccountId }/balances`, {
    enabled: !!selectedBankAccountId,
  });
  return result?.data && new Balance(result?.data);
}
