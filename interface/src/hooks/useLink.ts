import { useQuery, useQueryClient, type UseQueryResult } from '@tanstack/react-query';

import Link from '@monetr/interface/models/Link';

export function useLink(linkId?: string): UseQueryResult<Link, unknown> {
  const queryClient = useQueryClient();
  return useQuery<Partial<Link>, unknown, Link>({
    queryKey: [`/links/${linkId}`],
    enabled: Boolean(linkId),
    select: data => new Link(data),
    initialData: () => queryClient.getQueryData<Array<Link>>(['/links'])?.find(item => item.linkId === linkId),
    initialDataUpdatedAt: () => queryClient.getQueryState(['/links'])?.dataUpdatedAt,
  });
}
