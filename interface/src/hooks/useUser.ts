import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import type { ID } from '@monetr/interface/models/ID';
import User from '@monetr/interface/models/User';
import type { WithJsonValues } from '@monetr/interface/util/json';

export function useUser(userId: ID<User> | null): UseQueryResult<User, unknown> {
  return useQuery<WithJsonValues<User>, unknown, User>({
    queryKey: [`/api/users/${userId}`],
    enabled: Boolean(userId),
    select: data => new User(data),
  });
}
