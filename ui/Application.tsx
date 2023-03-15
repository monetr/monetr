import React, { lazy } from 'react';
import { Backdrop, CircularProgress } from '@mui/material';

import { useAppConfigurationSink } from 'hooks/useAppConfiguration';
import { useAuthenticationSink } from 'hooks/useAuthentication';
import CenteredLogo from 'components/Logo/CenteredLogo';

const UnauthenticatedApplication = lazy(() => import('UnauthenticatedApplication'));
const BillingRequiredRouter = lazy(() => import('BillingRequiredRouter'));
const AuthenticatedApp = lazy(() => import('AuthenticatedApp'));

export default function Application(): JSX.Element {
  const { isLoading: isLoadingConfig, isError } = useAppConfigurationSink();
  const { isLoading: isLoadingAuth, result: { user, isActive } } = useAuthenticationSink();
  const isLoading = isLoadingAuth || isLoadingConfig;
  const isAuthenticated = !!user;

  if (isError) {
    return (
      <Backdrop open>
        <div className='w-full h-full flex items-center justify-center'>
          <div className='w-1/4'>
            <CenteredLogo />
            <p className='text-center text-white text-lg'>
              It looks like monetr is having some problems right now; we should be back online shortly.
            </p>
          </div>
        </div>
      </Backdrop>
    );
  }

  if (isLoading) {
    return (
      <Backdrop open={ true }>
        <CircularProgress color="primary" />
      </Backdrop>
    );
  }

  if (!isAuthenticated) {
    return <UnauthenticatedApplication />;
  }

  if (!isActive) {
    return <BillingRequiredRouter />;
  }

  return (
    <AuthenticatedApp />
  );
}

