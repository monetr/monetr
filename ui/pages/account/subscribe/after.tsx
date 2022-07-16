import React from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { CircularProgress, Typography } from '@mui/material';

import { Logo } from 'assets';
import { useAfterCheckout } from 'hooks/useAuthentication';
import useMountEffect from 'hooks/useMountEffect';

export default function AfterCheckoutPage(): JSX.Element {
  const { search } = useLocation();
  const navigate = useNavigate();

  const afterCheckout = useAfterCheckout();

  async function setupFromCheckout(): Promise<void> {
    const params = new URLSearchParams(search);
    const checkoutSessionId = params.get('session');
    return afterCheckout(checkoutSessionId)
      .then(result => {
        // If the user's subscription is now active then redirect them to the main view of the authenticated
        // application.
        if (result.isActive) {
          return navigate('/');
        }

        // Otherwise, dispaly the message from the result of the afterCheckout call.
        alert(result?.message || 'Subscription is not active');
      })
      .catch(() => alert('Unable to determine your subscription state, please contact support@monetr.app'));
  }

  // As soon as the component mounts, call setup from checkout to get the subscription sorted out.
  useMountEffect(() => void setupFromCheckout());

  return (
    <div className="flex items-center justify-center w-full h-full max-h-full">
      <div className="w-full p-10 xl:w-3/12 lg:w-5/12 md:w-2/3 sm:w-10/12 max-w-screen-sm sm:p-0">
        <div className="flex justify-center w-full mb-5">
          <img src={ Logo } className="w-1/3" />
        </div>
        <div className="w-full pt-2.5 pb-2.5">
          <Typography
            variant="h5"
            className="w-full text-center"
          >
            Getting your account setup...
          </Typography>
        </div>
        <div className="w-full pt-2.5 pb-2.5 flex justify-center">
          <CircularProgress />
        </div>
      </div>
    </div>
  );
};
