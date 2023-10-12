import { LoaderFunctionArgs } from 'react-router-dom';

import queryClient from 'client';
import { AuthenticationWrapper } from 'hooks/useAuthentication';

export default function authenticationLoader(args: LoaderFunctionArgs): Promise<AuthenticationWrapper> {
  return queryClient.fetchQuery(['/users/me'], {
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}
