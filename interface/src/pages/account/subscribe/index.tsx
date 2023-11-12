import React, { useCallback, useState } from 'react';
import { useSnackbar } from 'notistack';

import { MBaseButton } from '@monetr/interface/components/MButton';
import MLink from '@monetr/interface/components/MLink';
import MLogo from '@monetr/interface/components/MLogo';
import MSpan from '@monetr/interface/components/MSpan';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import { useAuthenticationSink } from '@monetr/interface/hooks/useAuthentication';
import request from '@monetr/interface/util/request';

export default function SubscribePage(): JSX.Element {
  const { enqueueSnackbar } = useSnackbar();
  const {
    initialPlan,
  } = useAppConfiguration();
  const { result: { hasSubscription, activeUntil } } = useAuthenticationSink();

  const [loading, setLoading] = useState(false);
  const handleContinue = useCallback(() => {
    setLoading(true);
    let promise: Promise<any>;
    if (initialPlan && !hasSubscription) {
      promise = request().post('/billing/create_checkout', {
        priceId: '',
        cancelPath: '/logout',
      });
    } else if (hasSubscription) {
      // If the customer has a subscription then we want to just manage it. This will allow a customer to fix a
      // subscription for a card that has failed payment or something similar.
      promise = request().get('/billing/portal');
    }

    promise
      .then(result => window.location.assign(result.data.url))
      .catch(error => {
        setLoading(false);
        enqueueSnackbar(error?.response?.data?.error || 'Failed to prepare Stripe billing session.', {
          variant: 'error',
          disableWindowBlurListener: true,
        });
      });
  }, [enqueueSnackbar, hasSubscription, initialPlan]);

  if (activeUntil) {
    return (
      <div className='flex items-center justify-center w-full h-full max-h-full p-4'>
        <div className='h-full flex flex-col max-w-md gap-4 items-center justify-between'>
          <div className='h-full flex flex-col justify-center items-center gap-4'>
            <MLogo className='max-h-24' />
            <MSpan size='2xl' weight='semibold' className='text-center'>
              Your subscription is no longer active
            </MSpan>
            <MSpan size='lg' className='text-center'>
              Thank you for having subscribed to monetr before! If you'd like to continue using monetr you will have to
              resubscribe below. Click continue to proceed to our billing portal.
            </MSpan>
            <MBaseButton color='primary' disabled={ loading } onClick={ handleContinue }>
              Continue
            </MBaseButton>
          </div>
          { !loading &&
          <div className='flex justify-center gap-1'>
            <MSpan color='subtle' className='text-sm'>Not ready to continue?</MSpan>
            <MLink to='/logout' size='sm'>Logout for now</MLink>
          </div>
          }
        </div>
      </div>
    );
  }

  return (
    <div className='flex items-center justify-center w-full h-full max-h-full p-4'>
      <div className='h-full flex flex-col max-w-md gap-4 items-center justify-between'>
        <div className='h-full flex flex-col justify-center items-center gap-4'>
          <MLogo className='max-h-24' />
          <MSpan size='2xl' weight='semibold' className='text-center'>
            Your free trial has ended
          </MSpan>
          <MSpan size='lg' className='text-center'>
            Thank you for trying out monetr! We hope that you found our budgeting tools useful during your trial. If
            you'd like to continue using monetr you can easily subscribe below.
          </MSpan>
          <MBaseButton color='primary' disabled={ loading } onClick={ handleContinue }>
            Continue
          </MBaseButton>
        </div>
        { !loading &&
          <div className='flex justify-center gap-1'>
            <MSpan color='subtle' className='text-sm'>Not ready to continue?</MSpan>
            <MLink to='/logout' size='sm'>Logout for now</MLink>
          </div>
        }
      </div>
    </div>
  );
}


