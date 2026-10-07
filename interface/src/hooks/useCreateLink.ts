import { useMutation } from '@tanstack/react-query';

import Link from '@monetr/interface/models/Link';
import type { WithJsonValues } from '@monetr/interface/util/json';
import request from '@monetr/interface/util/request';

export interface CreateLinkRequest {
  institutionName: string;
  description?: string;
  lunchFlowLinkId?: string;
}

export function useCreateLink(): (_link: CreateLinkRequest) => Promise<Link> {
  const mutate = useMutation({
    mutationFn: async (newLink: CreateLinkRequest): Promise<Link> => {
      return await request<WithJsonValues<Link>>({ method: 'POST', url: '/api/links', data: newLink }).then(
        result => new Link(result?.data),
      );
    },
    onSuccess: (newLink: Link, _a, _b, context) =>
      Promise.all([
        context.client.setQueryData(
          ['GET', '/api/links'],
          // If the list hasn't been loaded yet then leave it alone, otherwise we'd cache a list with only this one in it
          (previous: Array<Partial<Link>> | undefined) => previous?.concat(newLink),
        ),
        context.client.setQueryData(['GET', `/api/links/${newLink.linkId}`], newLink),
      ]),
  });

  return mutate.mutateAsync;
}
