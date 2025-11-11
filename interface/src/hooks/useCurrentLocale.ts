import { useMemo } from 'react';

import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';
import { getLocale } from '@monetr/interface/util/locale';

export function useCurrentLocale(): string {
  const { data: me } = useAuthentication();
  return useMemo(() => me?.user?.account?.locale ?? getLocale(), [me]);
}
