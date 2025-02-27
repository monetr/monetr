import { useQuery, UseQueryResult } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/bankAccounts';
import Balance from '@monetr/interface/models/Balance';

export function useBalance(bankAccountId: string): Balance | null {
  const result = useQuery<Partial<Balance>>([`/bank_accounts/${ bankAccountId }/balances`]);
  return result?.data && new Balance(result?.data);
}

export function useCurrentBalanceOld(): Balance | null {
  const selectedBankAccountId = useSelectedBankAccountId();
  const result = useQuery<Partial<Balance>>([`/bank_accounts/${ selectedBankAccountId }/balances`], {
    enabled: !!selectedBankAccountId,
  });
  return result?.data && new Balance(result?.data);
}

export function useCurrentBalance(): UseQueryResult<Balance> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Partial<Balance>, unknown, Balance>([`/bank_accounts/${ selectedBankAccountId }/balances`], {
    enabled: !!selectedBankAccountId,
    select: data => new Balance(data),
  });
}
