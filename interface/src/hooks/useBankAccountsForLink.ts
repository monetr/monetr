import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import BankAccount from '@monetr/interface/models/BankAccount';
import type { WithJsonValues } from '@monetr/interface/util/json';

export function useBankAccountsForLink(linkId?: string): UseQueryResult<Array<BankAccount>, unknown> {
  return useQuery<Array<WithJsonValues<BankAccount>>, unknown, Array<BankAccount>>({
    queryKey: ['GET', '/api/bank_accounts', { link_id: linkId }],
    enabled: Boolean(linkId),
    select: data => (data || []).map(item => new BankAccount(item)),
  });
}
