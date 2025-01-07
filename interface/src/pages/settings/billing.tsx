import React, { useCallback, useState } from 'react';
import { useLocation } from 'react-router-dom';
import { AccessTimeOutlined } from '@mui/icons-material';
import { AxiosResponse } from 'axios';
import { format, isPast, isThisYear } from 'date-fns';
import { useSnackbar } from 'notistack';

import MBadge from '@monetr/interface/components/MBadge';
import { MBaseButton } from '@monetr/interface/components/MButton';
import MDivider from '@monetr/interface/components/MDivider';
import MSpan from '@monetr/interface/components/MSpan';
import { useAuthenticationSink } from '@monetr/interface/hooks/useAuthentication';
import request from '@monetr/interface/util/request';

export default function SettingsBilling(): JSX.Element {
  const location = useLocation();
  const { enqueueSnackbar } = useSnackbar();
  const [loading, setLoading] = useState(false);
  const { result: { hasSubscription } } = useAuthenticationSink();
  const handleManageSubscription = useCallback(async () => {
    setLoading(true);
    let promise: Promise<AxiosResponse<{ url: string }>>;
    if (!hasSubscription) {
      promise = request().post('/billing/create_checkout', {
        // If the user backs out of the stripe checkout then return them to the current URL.
        cancelPath: location.pathname,
      });
    } else {
      // If the customer has a subscription then we want to just manage it. This will allow a customer to fix a
      // subscription for a card that has failed payment or something similar.
      promise = request().get('/billing/portal');
    }

    await promise
      .then(result => window.location.assign(result.data.url))
      .catch(error => {
        setLoading(false);
        enqueueSnackbar(error?.response?.data?.error || 'Failed to prepare Stripe billing session.', {
          variant: 'error',
          disableWindowBlurListener: true,
        });
      });
  }, [enqueueSnackbar, hasSubscription, location]);

  const manageSubscriptionText = hasSubscription ? 'Manage Your Subscription' : 'Subscribe Early';

  return (
    <div className='w-full flex flex-col p-4 max-w-xl'>
      <MSpan size='2xl' weight='bold' color='emphasis' className='mb-4'>
        Billing
      </MSpan>
      <MDivider />

      <div className='flex justify-between py-4'>
        <MSpan>
          Subscription Status
        </MSpan>
        <SubscriptionStatusBadge />
      </div>
      <MDivider />

      <MBaseButton
        className='ml-auto mt-4 max-w-xs'
        color='primary'
        disabled={ loading }
        onClick={ handleManageSubscription }
      >
        { manageSubscriptionText }
      </MBaseButton>
    </div>
  );
}

function SubscriptionStatusBadge(): JSX.Element {
  const { result: { hasSubscription, trialingUntil } } = useAuthenticationSink();

  if (hasSubscription) {
    return (
      <MBadge className='bg-green-600'>
        Active
      </MBadge>
    );
  }

  if (trialingUntil && isPast(trialingUntil)) {
    return (
      <MBadge className='bg-red-600'>
        Trial Expired
      </MBadge>
    );
  }

  if (trialingUntil) {
    const trialEndDate = isThisYear(trialingUntil) ?
      format(trialingUntil, 'MMMM do') :
      format(trialingUntil, 'MMMM do, yyyy');

    return (
      <MBadge className='bg-yellow-600'>
        <AccessTimeOutlined />
        Trialing Until { trialEndDate }
      </MBadge>
    );
  }

  return (
    <MBadge className='bg-pink-600'>
      Unknown
    </MBadge>
  );
}

