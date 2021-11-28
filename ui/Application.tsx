import { Backdrop, CircularProgress } from '@mui/material';
import AuthenticatedApp from 'AuthenticatedApp';
import React, { useEffect } from 'react';
import { useState } from 'react';
import { useSelector } from 'react-redux';
import useBootstrapLogin from 'shared/authentication/actions/bootstrapLogin';
import { getIsAuthenticated, getSubscriptionIsActive } from 'shared/authentication/selectors';
import useBootstrapApplication from 'shared/bootstrap/actions/bootstrapApplication';
import { getIsBootstrapped } from 'shared/bootstrap/selectors';
import UnauthenticatedApplication from 'UnauthenticatedApplication';
import BillingRequiredRouter from 'views/Subscriptions/BillingRequiredRouter';
import * as Sentry from '@sentry/react';

const Application = (): JSX.Element => {
  const [loading, setLoading] = useState(true);
  const isReady = useSelector(getIsBootstrapped);
  const isAuthenticated = useSelector(getIsAuthenticated);
  const isSubscriptionActive = useSelector(getSubscriptionIsActive);
  const bootstrapApplication = useBootstrapApplication();
  const bootstrapLogin = useBootstrapLogin();

  // I really only want this to run one time, isReady will only be false when the application is initially loading in
  // a user's web browser. So we can use this effect to see if we have already bootstrapped things. If we have not then
  // we can kick that process off here at the highest level of the application.
  useEffect(() => {
    if (!isReady) {
      const transaction = Sentry.startTransaction({ name: 'Bootstrapping Monetr' });
      bootstrapApplication()
        .then(() => bootstrapLogin())
        .catch(error => {
          throw error; // TODO Add something to handle this error.
        })
        .finally(() => {
          setLoading(false);
          transaction.finish();
        });
    }
  }, [isReady]);

  // When the application is still getting ready we want to just show a loading state to the user.
  if (!isReady || loading) {
    return (
      <Backdrop open={ true }>
        <CircularProgress color="inherit"/>
      </Backdrop>
    );
  }

  if (!isAuthenticated) {
    return <UnauthenticatedApplication/>
  }

  if (!isSubscriptionActive) {
    return <BillingRequiredRouter/>;
  }

  return (
    <AuthenticatedApp/>
  )
};

export default Application;