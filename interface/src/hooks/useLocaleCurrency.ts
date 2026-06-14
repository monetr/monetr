import { useCallback, useMemo } from 'react';
import type { UseQueryResult } from '@tanstack/react-query';

import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';
import { useCurrentLocale } from '@monetr/interface/hooks/useCurrentLocale';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import { type AmountType, amountToFriendly, formatAmount, friendlyToAmount } from '@monetr/interface/util/amounts';
import { fallback } from '@monetr/interface/util/fallback';

enum CurrencySource {
  UserDefault,
  BankAccount,
}

export interface LocaleCurrency {
  source: CurrencySource;
  locale: string;
  currency: string;
  friendlyToAmount: (value: number) => number;
  amountToFriendly: (value: number) => number;
  formatAmount: (value: number, kind: AmountType, signDisplay?: boolean) => string;
}

export const DefaultCurrency = 'USD';

/**
 * useLocaleCurrency takes an optional currency code, if the code is provided then it will return the locale currency
 * for that specific code. Otherwise it will use a default chain for the currency + locale to be returned.
 * It will use the current bank accounts currency (if there is one), then the user's default currency then the global
 * default currency.
 */
export default function useLocaleCurrency(forceCurrency?: string): UseQueryResult<LocaleCurrency, unknown> {
  const { data: me, ...authenticationState } = useAuthentication();
  const bankAccount = useSelectedBankAccount();
  const locale = useCurrentLocale();
  const currency = useMemo(() => {
    // Return the first defined currency, DefaultCurrency is always last so there is always something to fall back to.
    return fallback(forceCurrency, bankAccount?.data?.currency, me?.defaultCurrency, DefaultCurrency);
  }, [forceCurrency, me, bankAccount]);

  const friendlyToAmountCallback = useCallback(
    (value: number) => {
      return friendlyToAmount(value, locale, currency);
    },
    [locale, currency],
  );

  const amountToFriendlyCallback = useCallback(
    (value: number) => {
      return amountToFriendly(value, locale, currency);
    },
    [locale, currency],
  );

  const formatAmountCallback = useCallback(
    (value: number, kind: AmountType, signDisplay?: boolean): string => {
      return formatAmount(value, kind, locale, currency, signDisplay);
    },
    [locale, currency],
  );

  return {
    ...(bankAccount as any),
    ...(authenticationState as any),
    data: {
      source: bankAccount?.data ? CurrencySource.BankAccount : CurrencySource.UserDefault,
      locale: locale,
      currency: currency,
      friendlyToAmount: friendlyToAmountCallback,
      amountToFriendly: amountToFriendlyCallback,
      formatAmount: formatAmountCallback,
    },
  };
}
