import { useQuery, UseQueryResult } from '@tanstack/react-query';

import { useAuthenticationSink } from '@monetr/interface/hooks/useAuthentication';
import Link from '@monetr/interface/models/Link';

export function useLinks(): UseQueryResult<Array<Link>, unknown> {
  const { result: { user, isActive, mfaPending } } = useAuthenticationSink();
  return useQuery<Array<Partial<Link>>, unknown, Array<Link>>({
    queryKey: ['/links'],
    // Only request links if there is an authenticated user.
    enabled: !!user && isActive && !mfaPending,
    select: data => {
      if (Array.isArray(data)) {
        return data.map(item => new Link(item));
      }

      return [];
    },
  });
}
