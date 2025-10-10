import { useQuery, UseQueryResult } from '@tanstack/react-query';

import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';

export function useInstalledCurrencies(): UseQueryResult<Array<string>> {
  const { data } = useAuthentication();
  return useQuery<Array<string>>({
    queryKey: ['/locale/currency'],
    // Only allowed to fetch currency and locale information if we are authenticated.
    enabled: Boolean(data?.user),
  });
}
