import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import type BankAccount from '@monetr/interface/models/BankAccount';
import type { ID } from '@monetr/interface/models/ID';
import Spending from '@monetr/interface/models/Spending';
import type { WithJsonValues } from '@monetr/interface/util/json';

export function useSpendings(bankAccountId: ID<BankAccount>): UseQueryResult<Array<Spending>, unknown> {
  return useQuery<Array<WithJsonValues<Spending>>, unknown, Array<Spending>>({
    queryKey: [`/api/bank_accounts/${bankAccountId}/spending`],
    select: data => (data || []).map(item => new Spending(item)),
  });
}
