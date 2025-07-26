import { useQuery, UseQueryResult } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/bankAccounts';
import Balance from '@monetr/interface/models/Balance';

export function useCurrentBalance(): UseQueryResult<Balance> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Partial<Balance>, unknown, Balance>(
    [`/bank_accounts/${ selectedBankAccountId }/balances`], 
    {
      enabled: !!selectedBankAccountId,
      select: data => new Balance(data),
    },
  );
}
