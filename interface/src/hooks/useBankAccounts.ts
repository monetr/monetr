import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import { useLinks } from '@monetr/interface/hooks/useLinks';
import BankAccount from '@monetr/interface/models/BankAccount';
import type { WithJsonValues } from '@monetr/interface/util/json';

export function useBankAccounts(): UseQueryResult<Array<BankAccount>, unknown> {
  const { data: links } = useLinks();
  return useQuery<Array<WithJsonValues<BankAccount>>, unknown, Array<BankAccount>>({
    queryKey: ['/api/bank_accounts'],
    enabled: !!links && links.length > 0,
    select: data => data.map(item => new BankAccount(item)),
  });
}
