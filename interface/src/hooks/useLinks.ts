import { useQuery, UseQueryResult } from '@tanstack/react-query';

import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';
import Link from '@monetr/interface/models/Link';

export function useLinks(): UseQueryResult<Array<Link>, unknown> {
  const { data: { user, isActive, mfaPending } } = useAuthentication();
  return useQuery<Array<Partial<Link>>, unknown, Array<Link>>({
    queryKey: ['/links'],
    // Only request links if there is an authenticated user.
    enabled: Boolean(user) && isActive && !mfaPending,
    initialData: [],
    initialDataUpdatedAt: 0,
    select: data => {
      if (Array.isArray(data)) {
        return data.map(item => new Link(item));
      }

      return [];
    },
  });
}
