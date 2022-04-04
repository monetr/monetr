import { CircularProgress, Typography } from '@mui/material';
import React from 'react';
import { getHasSubscription } from 'shared/authentication/selectors';
import request from 'shared/util/request';
import { useSelector } from 'react-redux';
import { getInitialPlan } from 'shared/bootstrap/selectors';
import useMountEffect from 'shared/util/useMountEffect';
import Logo from 'assets';

export default function SubscribePage(): JSX.Element {
  const initialPlan = useSelector(getInitialPlan);
  const hasSubscription = useSelector(getHasSubscription);

  useMountEffect(() => {
    if (initialPlan && !hasSubscription) {
      request().post(`/billing/create_checkout`, {
        priceId: '',
        cancelPath: '/logout',
      })
        .then(result => window.location.assign(result.data.url))
        .catch(error => alert(error));
    } else if (hasSubscription) {
      // If the customer has a subscription then we want to just manage it. This will allow a customer to fix a
      // subscription for a card that has failed payment or something similar.
      request().get('/billing/portal')
        .then(result => window.location.assign(result.data.url))
        .catch(error => alert(error));
    }
  });

  return (
    <div className="flex items-center justify-center w-full h-full max-h-full">
      <div className="w-full p-10 xl:w-3/12 lg:w-5/12 md:w-2/3 sm:w-10/12 max-w-screen-sm sm:p-0">
        <div className="flex justify-center w-full mb-5">
          {/*<img src={ Logo } className="w-1/3"/>*/ }
        </div>
        <div className="w-full pt-2.5 pb-2.5">
          <Typography
            variant="h5"
            className="w-full text-center"
          >
            Getting Stripe ready...
          </Typography>
        </div>
        <div className="w-full pt-2.5 pb-2.5 flex justify-center">
          <CircularProgress/>
        </div>
      </div>
    </div>
  );
}


