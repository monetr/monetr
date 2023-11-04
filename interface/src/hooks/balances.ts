import { useQuery } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/bankAccounts';
import Balance from '@monetr/interface/models/Balance';

export function useBalance(bankAccountId: number): Balance | null {
  const result = useQuery<Partial<Balance>>([`/bank_accounts/${ bankAccountId }/balances`]);
  return result?.data && new Balance(result?.data);
}

export function useCurrentBalance(): Balance | null {
  const selectedBankAccountId = useSelectedBankAccountId();
  const result = useQuery<Partial<Balance>>([`/bank_accounts/${ selectedBankAccountId }/balances`], {
    enabled: !!selectedBankAccountId,
  });
  return result?.data && new Balance(result?.data);
}
