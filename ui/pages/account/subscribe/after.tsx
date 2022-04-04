import React, { useEffect } from 'react';
import { useDispatch } from 'react-redux';
import { useLocation, useNavigate } from 'react-router-dom';
import request from 'shared/util/request';
import { CircularProgress, Typography } from '@mui/material';
import activateSubscription from 'shared/authentication/actions/activateSubscription';

import Logo from 'assets';

export default function AfterCheckoutPage(): JSX.Element {
  const { search } = useLocation();
  const dispatch = useDispatch();
  const navigate = useNavigate();

  function setupFromCheckout(): Promise<void> {
    const params = new URLSearchParams(search);
    const checkoutSessionId = params.get('session');
    return request().get(`/billing/checkout/${ checkoutSessionId }`)
      .then(({ data }) => {
        if (data?.isActive) {
          dispatch(activateSubscription());
          return navigate('/');
        }

        alert('subscription is not active');
      });
  }

  // As soon as the component mounts, call setup from checkout to get the subscription sorted out.
  useEffect(() => void setupFromCheckout(), []);

  return (
    <div className="flex items-center justify-center w-full h-full max-h-full">
      <div className="w-full p-10 xl:w-3/12 lg:w-5/12 md:w-2/3 sm:w-10/12 max-w-screen-sm sm:p-0">
        <div className="flex justify-center w-full mb-5">
          <img src={ Logo } className="w-1/3"/>
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
          <CircularProgress/>
        </div>
      </div>
    </div>
  );
};
