import { useCallback, useMemo } from 'react';
import { UseQueryResult } from '@tanstack/react-query';

import { useSelectedBankAccount } from '@monetr/interface/hooks/bankAccounts';
import { useAuthenticationSink } from '@monetr/interface/hooks/useAuthentication';
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

export default function useLocaleCurrency(): UseQueryResult<LocaleCurrency | null> {
  const { result: _, ...me } = useAuthenticationSink();
  const bankAccount = useSelectedBankAccount();
  const locale = useMemo(() => me.data?.user?.account?.locale ?? 'en_US', [me]);
  const currency = useMemo(() => {
    return Boolean(bankAccount.data) ? bankAccount.data.currency : (me.data?.defaultCurrency ?? 'USD');
  }, [me, bankAccount]);

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
    ...me as any,
    data: {
      source: Boolean(bankAccount.data) ? CurrencySource.BankAccount : CurrencySource.UserDefault,
      locale: locale,
      currency: currency,
      friendlyToAmount: friendlyToAmountCallback,
      amountToFriendly: amountToFriendlyCallback,
      formatAmount: formatAmountCallback,
    },
  };
}
