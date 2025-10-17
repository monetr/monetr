import { useQuery, UseQueryResult } from '@tanstack/react-query';

import { QueryMethod } from '@monetr/interface/components/MQueryClient';
import BankAccount from '@monetr/interface/models/BankAccount';

export function useBankAccountsForLink(linkId?: string): UseQueryResult<Array<BankAccount>, unknown> {
  return useQuery<Array<Partial<BankAccount>>, unknown, Array<BankAccount>>({
    queryKey: ['/bank_accounts', { link_id: linkId }],
    enabled: Boolean(linkId),
    meta: {
      method: QueryMethod.UseQuery,
    },
    select: data => (data || []).map(item => new BankAccount(item)),
  });
}
