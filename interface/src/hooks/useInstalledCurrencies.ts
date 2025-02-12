import { useQuery, UseQueryResult } from '@tanstack/react-query';

import { useAuthenticationSink } from '@monetr/interface/hooks/useAuthentication';

export function useInstalledCurrencies(): UseQueryResult<Array<string>> {
  const { result } = useAuthenticationSink();
  return useQuery<Array<string>>(
    ['/locale/currency'],
    { 
      // Only allowed to fetch currency and locale information if we are authenticated.
      enabled: !!(result?.user),
    }
  );
}
