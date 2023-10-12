import { redirect } from 'react-router-dom';

import queryClient from 'client';
import { AuthenticationWrapper } from 'hooks/useAuthentication';

function getMe(): Promise<Partial<AuthenticationWrapper>> {
  return queryClient.fetchQuery<Partial<AuthenticationWrapper>>(['/users/me'], {
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

export async function requiresNotAuthenticated(): Promise<null | Response> {
  console.trace('checking authentication (required absent)');
  return getMe().then(result => {
    if (result?.user?.userId) {
      // If the user ID is present then this route is not valid. Redirect to the index route.
      return redirect('/');
    }

    return null;
  });
}

export async function requiresAuthentication(): Promise<null | Response> {
  console.trace('checking authentication (required present)');
  return getMe().then(result => {
    if (!result?.user?.userId) {
      // If the user ID is not present then redirect to login.
      return redirect('/login');
    }

    return null;
  });
}

export async function indexLoader(): Promise<Response> {
  return getMe().then(result => {
    if (!result?.user?.userId) {
      // If the user ID is present then this route is not valid. Redirect to the index route.
      return redirect('/login');
    }

    if (!result?.isActive) {
      return redirect('/account/subscribe');
    }

    if (!result?.isSetup) {
      return redirect('/setup');
    }

    console.error('TODO, redirect to the correct authenticated route');
    return null;
  });
}

