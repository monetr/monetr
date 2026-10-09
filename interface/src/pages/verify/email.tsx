import { useCallback, useEffect } from 'react';
import { useLocation, useSearch } from 'wouter';

import MLogo from '@monetr/interface/components/MLogo';
import Typography from '@monetr/interface/components/Typography';
import request from '@monetr/interface/util/request';
import { useSnackbar, type VariantType } from '@monetr/notify';

import styles from './email.module.scss';

export default function VerifyEmail(): React.JSX.Element {
  const search = useSearch();
  const [, navigate] = useLocation();
  const { enqueueSnackbar } = useSnackbar();

  const errorRedirect = useCallback(
    (message: string, variant: VariantType, nextUrl: string = '/login') => {
      enqueueSnackbar(message, {
        variant: variant,
        disableWindowBlurListener: true,
      });
      navigate(nextUrl);
    },
    [enqueueSnackbar, navigate],
  );

  const query = new URLSearchParams(search);
  const token = query.get('token');

  useEffect(() => {
    if (!token) {
      errorRedirect('Email verification link is not valid.', 'error');
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
          'success',
          result?.data?.nextUrl || '/login',
        ),
      )
      .catch(error =>
        errorRedirect(
          error?.response?.data?.error || 'Failed to verify email address.',
          'error',
          error?.response?.data?.nextUrl,
        ),
      );
  }, [token, errorRedirect]);

  return <VerifyEmailView />;
}

export function VerifyEmailView(): React.JSX.Element {
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
