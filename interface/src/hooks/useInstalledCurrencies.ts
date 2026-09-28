import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';

export interface Currency {
  code: string;
  name: string;
  symbol: string;
  fractionalDigits: number;
}

export function useInstalledCurrencies(): UseQueryResult<Array<Currency>> {
  const { data } = useAuthentication();
  return useQuery<Array<Currency>>({
    queryKey: ['/api/locale/currency'],
    // Only allowed to fetch currency and locale information if we are authenticated.
    enabled: Boolean(data?.user),
  });
}
