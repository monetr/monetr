import { useCallback } from 'react';
import { type UseQueryResult, useQuery, useQueryClient } from '@tanstack/react-query';

import Link from '@monetr/interface/models/Link';
import type { WithJsonValues } from '@monetr/interface/util/json';

export function useLink(linkId?: string): UseQueryResult<Link, unknown> {
  const queryClient = useQueryClient();
  const initialData = useCallback(
    () => queryClient.getQueryData<Array<Link>>(['/api/links'])?.find(item => item.linkId === linkId),
    [queryClient, linkId],
  );
  const initialDataUpdatedAt = useCallback(
    () => queryClient.getQueryState(['/api/links'])?.dataUpdatedAt,
    [queryClient],
  );
  const select = useCallback((data: WithJsonValues<Link>) => new Link(data), []);
  return useQuery<WithJsonValues<Link>, unknown, Link>({
    queryKey: [`/api/links/${linkId}`],
    enabled: Boolean(linkId),
    select,
    initialData: initialData,
    initialDataUpdatedAt: initialDataUpdatedAt,
    notifyOnChangeProps: ['data', 'isLoading', 'isError'],
  });
}
