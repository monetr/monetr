import React from 'react';
import { Backdrop, CircularProgress } from '@mui/material';

import AuthenticatedApp from 'AuthenticatedApp';
import BillingRequiredRouter from 'BillingRequiredRouter';
import { useAppConfigurationSink } from 'hooks/useAppConfiguration';
import { useAuthenticationSink } from 'hooks/useAuthentication';
import UnauthenticatedApplication from 'UnauthenticatedApplication';

const Application = (): JSX.Element => {
  const { isLoading, isError } = useAppConfigurationSink();
  const { result: { user, isActive } } = useAuthenticationSink();
  const isReady = !isLoading && !isError;
  const isAuthenticated = !!user;

  // When the application is still getting ready we want to just show a loading state to the user.
  if (!isReady) {
    return (
      <Backdrop open={ true }>
        <CircularProgress color="inherit" />
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
};

export default Application;
