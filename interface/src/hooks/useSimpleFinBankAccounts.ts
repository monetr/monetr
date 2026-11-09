import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import SimpleFinBankAccount from '@monetr/interface/models/SimpleFinBankAccount';
import type { WithJsonValues } from '@monetr/interface/util/json';

export function useSimpleFinBankAccounts(
  simpleFinLinkId?: string,
): UseQueryResult<Array<SimpleFinBankAccount>, unknown> {
  return useQuery<Array<WithJsonValues<SimpleFinBankAccount>>, unknown, Array<SimpleFinBankAccount>>({
    queryKey: ['GET', `/api/simplefin/link/${simpleFinLinkId}/bank_accounts`],
    enabled: Boolean(simpleFinLinkId),
    select: data => (data ?? []).map(item => new SimpleFinBankAccount(item)),
  });
}
