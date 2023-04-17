import React from 'react';
import { CircularProgress } from '@mui/material';

import AuthenticatedApp from 'AuthenticatedApp';
import BillingRequiredRouter from 'BillingRequiredRouter';
import { useAppConfigurationSink } from 'hooks/useAppConfiguration';
import { useAuthenticationSink } from 'hooks/useAuthentication';
import UnauthenticatedApplication from 'UnauthenticatedApplication';
import CenteredLogo from 'components/Logo/CenteredLogo';

export default function Application(): JSX.Element {
  const { isLoading: isLoadingConfig, isError } = useAppConfigurationSink();
  const { isLoading: isLoadingAuth, result: { user, isActive } } = useAuthenticationSink();
  const isLoading = isLoadingAuth || isLoadingConfig;
  const isAuthenticated = !!user;

  if (isError) {
    return (
      <div className='w-full h-full flex flex-col items-center justify-center gap-y-4'>
        <div className='w-1/4'>
          <CenteredLogo />
        </div>
        <p className='text-center dark:text-white text-3xl'>
          Something isn't quite right...
        </p>
        <p className='text-center dark:text-white text-lg'>
          It looks like monetr is having some problems right now; we should be back online shortly.
        </p>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-y-4'>
        <div className='w-1/4'>
          <CenteredLogo />
        </div>
        <CircularProgress color="secondary" />
        <p className='text-center text-3xl dark:text-white'>One moment...</p>
      </div>
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

