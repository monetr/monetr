import { useCallback, useEffect } from 'react';
import { useLocation, useSearch } from 'wouter';

import MLogo from '@monetr/interface/components/MLogo';
import Typography from '@monetr/interface/components/Typography';
import request from '@monetr/interface/util/request';

import styles from './email.module.scss';

export default function VerifyEmail(): JSX.Element {
  const search = useSearch();
  const [, navigate] = useLocation();

  const errorRedirect = useCallback(
    (message: string, nextUrl: string = '/login') => {
      window.alert(message);
      navigate(nextUrl);
    },
    [navigate],
  );

  const query = new URLSearchParams(search);
  const token = query.get('token');

  useEffect(() => {
    if (!token) {
      errorRedirect('Email verification link is not valid.');
      return;
    }

    request<{ message?: string; nextUrl?: string }>({
      method: 'POST',
      url: '/api/authentication/verify',
      data: {
        token: token,
      },
    })
      .then(result =>
        errorRedirect(
          result?.data?.message || 'Your email has been verified, please login.',
          result?.data?.nextUrl || '/login',
        ),
      )
      .catch(error =>
        errorRedirect(
          error?.response?.data?.error || 'Failed to verify email address.',
          error?.response?.data?.nextUrl,
        ),
      );
  }, [token, errorRedirect]);

  return <VerifyEmailView />;
}

export function VerifyEmailView(): JSX.Element {
  return (
    <div className={styles.root}>
      <MLogo className={styles.logo} />
      <Typography size='2xl' weight='bold'>
        Email Verification
      </Typography>
      <Typography align='center' size='xl'>
        Your email is being verified, one moment...
      </Typography>
    </div>
  );
}
