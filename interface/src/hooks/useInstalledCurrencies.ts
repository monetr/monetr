import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';

export interface Currency {
  /** ISO 4217 currency code, like USD. */
  code: string;
  /** Name of the currency localized to the browser's language, like "US Dollar". */
  name: string;
  /** Symbol for the currency localized to the browser's language, like "$" or "US$". */
  symbol: string;
  /** Number of digits after the decimal place, USD uses 2 and JPY uses 0. */
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
