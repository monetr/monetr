import { useCallback } from 'react';
import { type UseQueryResult, useQuery, useQueryClient } from '@tanstack/react-query';

import Link from '@monetr/interface/models/Link';

export function useLink(linkId?: string): UseQueryResult<Link, unknown> {
  const queryClient = useQueryClient();
  const initialData = useCallback(
    () => queryClient.getQueryData<Array<Link>>(['/links'])?.find(item => item.linkId === linkId),
    [queryClient, linkId],
  );
  const initialDataUpdatedAt = useCallback(() => queryClient.getQueryState(['/links'])?.dataUpdatedAt, [queryClient]);
  const select = useCallback((data: Partial<Link>) => new Link(data), []);
  return useQuery<Partial<Link>, unknown, Link>({
    queryKey: [`/links/${linkId}`],
    enabled: Boolean(linkId),
    select,
    initialData: initialData,
    initialDataUpdatedAt: initialDataUpdatedAt,
    notifyOnChangeProps: ['data', 'isLoading', 'isError'],
  });
}
