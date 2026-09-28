import { useCallback } from 'react';
import { type UseQueryResult, useQuery, useQueryClient } from '@tanstack/react-query';

import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';
import type { Currency } from '@monetr/interface/hooks/useInstalledCurrencies';

/**
 * useCurrency retrieves the details of a single currency by its ISO 4217 code, if we already have the full currency
 * list then that gets used as the initial data
 */
export function useCurrency(code?: string): UseQueryResult<Currency> {
  const { data } = useAuthentication();
  const queryClient = useQueryClient();
  const initialData = useCallback(
    () => queryClient.getQueryData<Array<Currency>>(['/api/locale/currency'])?.find(item => item.code === code),
    [queryClient, code],
  );
  const initialDataUpdatedAt = useCallback(
    () => queryClient.getQueryState(['/api/locale/currency'])?.dataUpdatedAt,
    [queryClient],
  );
  return useQuery<Currency>({
    queryKey: [`/api/locale/currency/${code}`],
    // Only allowed to fetch currency and locale information if we are authenticated
    enabled: Boolean(data?.user) && Boolean(code),
    initialData,
    initialDataUpdatedAt,
  });
}
