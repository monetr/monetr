import { useMutation, useQueryClient } from '@tanstack/react-query';

import Link from '@monetr/interface/models/Link';
import request from '@monetr/interface/util/request';

export interface CreateLinkRequest {
  institutionName: string;
  description?: string;
}

export function useCreateLink(): (_link: CreateLinkRequest) => Promise<Link> {
  const queryClient = useQueryClient();

  async function createLink(newLink: CreateLinkRequest): Promise<Link> {
    return request()
      .post<Partial<Link>>('/links', newLink)
      .then(result => new Link(result?.data));
  }

  const mutate = useMutation(
    {
      mutationFn: createLink,
      onSuccess: (newLink: Link) => Promise.all([
        queryClient.setQueryData(
          ['/links'],
          (previous: Array<Partial<Link>> | null) => (previous ?? []).concat(newLink),
        ),
        queryClient.setQueryData(
          [`/links/${newLink.linkId}`],
          newLink,
        ),
      ]),
    }
  );

  return mutate.mutateAsync;
}
