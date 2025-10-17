import { useMutation, useQueryClient } from '@tanstack/react-query';

import Link from '@monetr/interface/models/Link';
import request from '@monetr/interface/util/request';

export type PatchLinkRequest = Pick<Link, 'linkId'> & Partial<Pick<Link, 'institutionName' | 'description'>>;

export function usePatchLink(): (_: PatchLinkRequest) => Promise<Link> {
  const queryClient = useQueryClient();

  async function patchLink({ linkId, ...patch }: PatchLinkRequest): Promise<Link> {
    return request()
      .patch(`/links/${linkId}`, patch)
      .then(result => new Link(result.data));
  }

  const mutation = useMutation({
    mutationFn: patchLink,
    onSuccess: (updated: Link) =>
      Promise.all([
        queryClient.setQueryData(['/links'], (previous: Array<Partial<Link>>) =>
          (previous ?? [])
            // Take the existing data in the cache and map over it, returning the updated item instead of the old item.
            .map(existing => (existing.linkId === updated.linkId ? updated : existing)),
        ),
        queryClient.setQueryData([`/links/${updated.linkId}`], updated),
      ]),
  });

  return mutation.mutateAsync;
}
