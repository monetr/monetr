import { useMutation } from '@tanstack/react-query';

import Link from '@monetr/interface/models/Link';
import request from '@monetr/interface/util/request';

export interface CreateLinkRequest {
  institutionName: string;
  description?: string;
}

export function useCreateLink(): (_link: CreateLinkRequest) => Promise<Link> {
  const mutate = useMutation({
    mutationFn: async (newLink: CreateLinkRequest): Promise<Link> => {
      return request()
        .post<Partial<Link>>('/links', newLink)
        .then(result => new Link(result?.data));
    },
    onSuccess: (newLink: Link, _a, _b, context) =>
      Promise.all([
        context.client.setQueryData(['/links'], (previous: Array<Partial<Link>> | null) =>
          (previous ?? []).concat(newLink),
        ),
        context.client.setQueryData([`/links/${newLink.linkId}`], newLink),
      ]),
  });

  return mutate.mutateAsync;
}
