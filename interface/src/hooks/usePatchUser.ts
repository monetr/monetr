import { useMutation } from '@tanstack/react-query';

import type { Authentication } from '@monetr/interface/hooks/useAuthentication';
import type { ID } from '@monetr/interface/models/ID';
import User from '@monetr/interface/models/User';
import type { WithJsonValues } from '@monetr/interface/util/json';
import type { Writable } from '@monetr/interface/util/readonly';
import request from '@monetr/interface/util/request';

export type PatchUserRequest = Partial<Writable<User>> & {
  userId: ID<User>;
};

export type PatchUserResponse = WithJsonValues<User>;

export function usePatchUser(): (patch: PatchUserRequest) => Promise<PatchUserResponse> {
  const { mutateAsync } = useMutation({
    async mutationFn({ userId, ...user }: PatchUserRequest): Promise<PatchUserResponse> {
      return await request<PatchUserResponse>({
        method: 'PATCH',
        url: `/api/users/${userId}`,
        data: user,
      }).then(result => result.data);
    },
    // Apply the patch to the cached /api/users/me before the request even goes out. Things like link order are read off
    // of the current user, so if we only invalidated after the request finished then a drag to reorder would snap back
    // to the old order until the refetch landed.
    onMutate: async ({ userId: _, ...patch }, { client: queryClient }) => {
      // /api/users/me refetches on window focus, cancel anything in flight so an older response can't land on top of
      // the optimistic value.
      await queryClient.cancelQueries({ queryKey: ['GET', '/api/users/me'] });
      const previous = queryClient.getQueryData<Partial<Authentication>>(['GET', '/api/users/me']);
      queryClient.setQueryData<Partial<Authentication>>(['GET', '/api/users/me'], current =>
        current?.user ? { ...current, user: new User({ ...current.user, ...patch }) } : current,
      );
      return { previous };
    },
    // If the server rejected the patch then put the cache back the way it was.
    onError: (_error, _input, onMutateResult, { client: queryClient }) =>
      queryClient.setQueryData(['GET', '/api/users/me'], onMutateResult?.previous),
    onSuccess: (result: PatchUserResponse, _input, _, { client: queryClient }) =>
      queryClient.setQueryData(['GET', `/api/users/${result.userId}`], result),
    // Refetch either way so the cache ends up matching the server. On success nothing visibly changes since the server
    // already has what we optimistically wrote.
    onSettled: (_data, _error, _input, _, { client: queryClient }) =>
      queryClient.invalidateQueries({ queryKey: ['GET', '/api/users/me'] }),
  });

  return mutateAsync;
}
