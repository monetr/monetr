import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import Balance from '@monetr/interface/models/Balance';
import type { WithJsonValues } from '@monetr/interface/util/json';

export function useCurrentBalance(): UseQueryResult<Balance, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<WithJsonValues<Balance>, unknown, Balance>({
    queryKey: [`/api/bank_accounts/${selectedBankAccountId}/balances`],
    enabled: !!selectedBankAccountId,
    select: data => new Balance(data),
  });
}
