import { useCallback, useMemo } from 'react';
import { UseQueryResult } from '@tanstack/react-query';

import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import { amountToFriendly, AmountType, formatAmount, friendlyToAmount } from '@monetr/interface/util/amounts';

enum CurrencySource {
  UserDefault,
  BankAccount,
}

interface LocaleCurrency {
  source: CurrencySource,
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
export default function useLocaleCurrency(forceCurrency?: string): UseQueryResult<LocaleCurrency> {
  const { data: me, ...authenticationState } = useAuthentication();
  const bankAccount = useSelectedBankAccount();
  const locale = useMemo(() => me?.user?.account?.locale ?? 'en_US', [me]);
  const currency = useMemo(() => {
    // Return the first _defined_ currency.
    return [
      forceCurrency,
      bankAccount?.data?.currency,
      me?.defaultCurrency,
      DefaultCurrency,
    ].find(value => !!value);
  }, [forceCurrency, me, bankAccount]);

  const friendlyToAmountCallback = useCallback((value: number) => {
    return friendlyToAmount(value, locale, currency);
  }, [locale, currency]);

  const amountToFriendlyCallback = useCallback((value: number) => {
    return amountToFriendly(value, locale, currency);
  }, [locale, currency]);

  const formatAmountCallback = useCallback((value: number, kind: AmountType, signDisplay?: boolean): string => {
    return formatAmount(value, kind, locale, currency, signDisplay);
  }, [locale, currency]);

  return {
    ...bankAccount as any,
    ...authenticationState as any,
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
