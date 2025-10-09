import { useEffect } from 'react';
import * as Sentry from '@sentry/react';
import { useQuery, UseQueryResult } from '@tanstack/react-query';

import { DefaultCurrency } from '@monetr/interface/hooks/useLocaleCurrency';
import User from '@monetr/interface/models/User';
import parseDate from '@monetr/interface/util/parseDate';

export interface Authentication {
  user: User;
  defaultCurrency: string;
  mfaPending: boolean;
  isSetup: boolean;
  isActive: boolean;
  isTrialing: boolean;
  activeUntil: Date | null;
  trialingUntil: Date | null;
  hasSubscription: boolean;
}

export function useAuthentication(): UseQueryResult<Authentication | undefined, unknown> {
  const result = useQuery<Partial<Authentication>, unknown, Authentication>({
    queryKey: ['/users/me'],
    select: data => ({
      user: Boolean(data?.user) && new User(data?.user),
      defaultCurrency: data?.defaultCurrency || DefaultCurrency,
      mfaPending: Boolean(data?.mfaPending),
      isSetup: Boolean(data?.isSetup),
      isActive: Boolean(data?.isActive),
      isTrialing: Boolean(data?.isTrialing),
      activeUntil: parseDate(data?.activeUntil),
      trialingUntil: parseDate(data?.trialingUntil),
      hasSubscription: Boolean(data?.hasSubscription),
    }),
    refetchOnWindowFocus: true, // Might want to change this to 'always' at some point?
  });

  // When we go from not being logged in, to being logged in; we should automatically set the user context for sentry!
  useEffect(() => {
    if (result?.data?.user?.accountId) {
      Sentry.setUser({
        id: result.data.user.accountId,
        username: `account:${result.data.user.accountId}`,
      });
    }
  }, [result]);

  return result;
}

