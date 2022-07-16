import { useQuery, UseQueryResult } from 'react-query';

import User from 'models/User';

export interface AuthenticationWrapper {
  user: User;
  isSetup: boolean;
  isActive: boolean;
  hasSubscription: boolean;
}

export type AuthenticationResult =
  { result: AuthenticationWrapper }
  & UseQueryResult<Partial<AuthenticationWrapper>, unknown>;

export function useAuthenticationSink(): AuthenticationResult {
  const result = useQuery<Partial<AuthenticationWrapper>>('/api/users/me');
  return {
    ...result,
    result: {
      user: result?.data?.user && new User(result?.data?.user),
      isSetup: !!result?.data?.isSetup,
      isActive: !!result?.data?.isActive,
      hasSubscription: !!result?.data?.hasSubscription,
    },
  };
}

export function useAuthentication(): User | null {
  const { result: { user } } = useAuthenticationSink();
  return user;
}
