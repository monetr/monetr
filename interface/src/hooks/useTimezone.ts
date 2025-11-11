import { useMemo } from 'react';
import { tz } from '@date-fns/tz';

import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';
import { getTimezone } from '@monetr/interface/util/locale';

export type TimezoneResult = {
  timezone: string;
  inTimezone: ReturnType<typeof tz>;
};

/**
 * useTimezone is the same or similar to the useAuthentication hook, however it will always return a timezone string no
 * matter what. It will always have initial data. If the user's timezone is not accessible via the API then it will be
 * derived from the browser's current timezone.
 */
export default function useTimezone(): TimezoneResult {
  const { data: me } = useAuthentication();
  return useMemo(() => {
    const timezone = me?.user?.account?.timezone ?? getTimezone();
    return {
      timezone,
      inTimezone: tz(timezone),
    };
  }, [me]);
}
