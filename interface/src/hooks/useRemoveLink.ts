import { useMutation, useQueryClient } from '@tanstack/react-query';

import Link from '@monetr/interface/models/Link';
import request from '@monetr/interface/util/request';

export function useRemoveLink(): (_linkId: string) => Promise<unknown> {
  const queryClient = useQueryClient();

  async function removeLink(linkId: string): Promise<string> {
    return request()
      .delete(`/links/${linkId}`)
      .then(() => linkId);
  }

  const mutate = useMutation({
    mutationFn: removeLink,
    onSuccess: (linkId: string) => Promise.all([
      queryClient.setQueryData(
        ['/links'],
        (previous: Array<Partial<Link>>) => previous.filter(item => item.linkId !== linkId),
      ),
      queryClient.removeQueries({ queryKey: [`/links/${linkId}`] }),
    ]),
  });

  return mutate.mutateAsync;
}
