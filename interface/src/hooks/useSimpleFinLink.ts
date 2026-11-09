import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import SimpleFinLink from '@monetr/interface/models/SimpleFinLink';
import type { WithJsonValues } from '@monetr/interface/util/json';

export function useSimpleFinLink(simpleFinLinkId?: string): UseQueryResult<SimpleFinLink, unknown> {
  return useQuery<WithJsonValues<SimpleFinLink>, unknown, SimpleFinLink>({
    queryKey: ['GET', `/api/simplefin/link/${simpleFinLinkId}`],
    enabled: Boolean(simpleFinLinkId),
    select: data => new SimpleFinLink(data),
  });
}
