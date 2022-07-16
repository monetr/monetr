import { useQuery } from 'react-query';

import { useBankAccountsSink } from 'hooks/bankAccounts';
import Balance from 'models/Balance';

export function useBalance(bankAccountId: number): Balance | null {
  const result = useQuery<Partial<Balance>>(`/api/bank_accounts/${ bankAccountId }/balances`);
  return result?.data && new Balance(result?.data);
}

export function useCurrentBalance(): Balance | null {
  const { result: { selectedBankAccountId } } = useBankAccountsSink();
  const result = useQuery<Partial<Balance>>(`/api/bank_accounts/${ selectedBankAccountId }/balances`, {
    enabled: !!selectedBankAccountId,
  });
  return result?.data && new Balance(result?.data);
}
