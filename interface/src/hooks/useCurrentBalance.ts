import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import Balance from '@monetr/interface/models/Balance';
import type BankAccount from '@monetr/interface/models/BankAccount';
import type { ID } from '@monetr/interface/models/ID';
import type { WithJsonValues } from '@monetr/interface/util/json';

export function useCurrentBalance(bankAccountId: ID<BankAccount>): UseQueryResult<Balance, unknown> {
  return useQuery<WithJsonValues<Balance>, unknown, Balance>({
    queryKey: [`/api/bank_accounts/${bankAccountId}/balances`],
    select: data => new Balance(data),
  });
}
