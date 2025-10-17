import { useQuery, UseQueryResult } from '@tanstack/react-query';

import { Authentication } from '@monetr/interface/hooks/useAuthentication';
import { getTimezone } from '@monetr/interface/util/locale';

/**
 * useTimezone is the same or similar to the useAuthentication hook, however it will always return a timezone string no
 * matter what. It will always have initial data. If the user's timezone is not accessible via the API then it will be
 * derived from the browser's current timezone.
 */
export default function useTimezone(): UseQueryResult<string, never> {
  return useQuery<Partial<Authentication>, never, string>({
    queryKey: ['/users/me'],
    initialData: () =>
      ({
        user: {
          account: {
            timezone: getTimezone(),
          },
        },
      }) as Partial<Authentication>,
    initialDataUpdatedAt: 0,
    select: data => data?.user?.account?.timezone ?? getTimezone(),
  });
}
