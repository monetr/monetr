import { useQuery, UseQueryResult } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import Balance from '@monetr/interface/models/Balance';

export function useCurrentBalance(): UseQueryResult<Balance, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Partial<Balance>, unknown, Balance>({
    queryKey: [`/bank_accounts/${selectedBankAccountId}/balances`],
    enabled: !!selectedBankAccountId,
    select: data => new Balance(data),
  });
}
