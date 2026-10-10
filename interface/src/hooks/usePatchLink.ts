import { useMutation } from '@tanstack/react-query';

import Link from '@monetr/interface/models/Link';
import type { WithJsonValues } from '@monetr/interface/util/json';
import request from '@monetr/interface/util/request';

export type PatchLinkRequest = Pick<Link, 'linkId'> &
  Partial<Pick<Link, 'institutionName' | 'description' | 'bankAccountOrder'>>;

export function usePatchLink(): (_: PatchLinkRequest) => Promise<Link> {
  const { mutateAsync } = useMutation({
    async mutationFn({ linkId, ...patch }: PatchLinkRequest): Promise<Link> {
      return await request<WithJsonValues<Link>>({ method: 'PATCH', url: `/api/links/${linkId}`, data: patch }).then(
        result => new Link(result.data),
      );
    },
    // Apply the patch to the cached links before the request even goes out. Things like the account order are read off
    // of the link, so if we only updated after the request finished then a drag to reorder would snap back to the old
    // order until the response landed.
    onMutate: async ({ linkId, ...patch }, { client: queryClient }) => {
      await Promise.all([
        queryClient.cancelQueries({ queryKey: ['GET', '/api/links'] }),
        queryClient.cancelQueries({ queryKey: ['GET', `/api/links/${linkId}`] }),
      ]);
      const previousLinks = queryClient.getQueryData<Array<Partial<Link>>>(['GET', '/api/links']);
      const previousLink = queryClient.getQueryData<Partial<Link>>(['GET', `/api/links/${linkId}`]);
      queryClient.setQueryData<Array<Partial<Link>>>(['GET', '/api/links'], current =>
        current?.map(existing => (existing.linkId === linkId ? { ...existing, ...patch } : existing)),
      );
      queryClient.setQueryData<Partial<Link>>(['GET', `/api/links/${linkId}`], current =>
        current ? { ...current, ...patch } : current,
      );
      return { previousLinks, previousLink };
    },
    // If the server rejected the patch then put the cache back the way it was.
    onError: (_error, { linkId }, onMutateResult, { client: queryClient }) => {
      queryClient.setQueryData(['GET', '/api/links'], onMutateResult?.previousLinks);
      queryClient.setQueryData(['GET', `/api/links/${linkId}`], onMutateResult?.previousLink);
    },
    onSuccess: (updated: Link, _input, _, { client: queryClient }) =>
      Promise.all([
        queryClient.setQueryData(
          ['GET', '/api/links'],
          // If the list hasn't been loaded yet then leave it alone, otherwise we'd cache an empty list
          (previous: Array<Partial<Link>> | undefined) =>
            // Take the existing data in the cache and map over it, returning the updated item instead of the old item.
            previous?.map(existing => (existing.linkId === updated.linkId ? updated : existing)),
        ),
        queryClient.setQueryData(['GET', `/api/links/${updated.linkId}`], updated),
      ]),
    // Refetch either way so the cache ends up matching the server. On success nothing visibly changes since the server
    // already has what we wrote.
    onSettled: (_data, _error, { linkId }, _, { client: queryClient }) =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: ['GET', '/api/links'] }),
        queryClient.invalidateQueries({ queryKey: ['GET', `/api/links/${linkId}`] }),
      ]),
  });

  return mutateAsync;
}
