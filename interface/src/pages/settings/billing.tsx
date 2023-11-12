import React, { Fragment, useCallback, useState } from 'react';
import { AccessTimeOutlined } from '@mui/icons-material';
import { format, isPast, isThisYear } from 'date-fns';
import { useSnackbar } from 'notistack';

import MBadge from '@monetr/interface/components/MBadge';
import { MBaseButton } from '@monetr/interface/components/MButton';
import MDivider from '@monetr/interface/components/MDivider';
import MSpan from '@monetr/interface/components/MSpan';
import { useAuthenticationSink } from '@monetr/interface/hooks/useAuthentication';
import request from '@monetr/interface/util/request';

export default function SettingsBilling(): JSX.Element {
  const { result: { trialingUntil } } = useAuthenticationSink();

  return (
    <div className='w-full flex flex-col p-4 max-w-xl'>
      <MSpan size='2xl' weight='bold' color='emphasis' className='mb-4'>
        Billing
      </MSpan>
      <MDivider />

      <TrialingRow trialingUntil={ trialingUntil } />
      <ActiveSubscriptionRow />
    </div>
  );
}

function ActiveSubscriptionRow(): JSX.Element {
  const { enqueueSnackbar } = useSnackbar();
  const [loading, setLoading] = useState(false);
  const { result: { hasSubscription } } = useAuthenticationSink();
  const handleManageSubscription = useCallback(() => {
    setLoading(true);
    // If the customer has a subscription then we want to just manage it. This will allow a customer to fix a
    // subscription for a card that has failed payment or something similar.
    request().get('/billing/portal')
      .then(result => window.location.assign(result.data.url))
      .catch(error => {
        setLoading(false);
        enqueueSnackbar(error?.response?.data?.error || 'Failed to prepare Stripe billing session.', {
          variant: 'error',
          disableWindowBlurListener: true,
        });
      });
  }, [enqueueSnackbar]);

  if (!hasSubscription) {
    return null;
  }

  return (
    <Fragment>
      <div className='flex justify-between py-4'>
        <MSpan>
          Subscription Status
        </MSpan>
        <MBadge className='bg-green-600'>
          Active
        </MBadge>
      </div>
      <MDivider />

      <MBaseButton
        className='ml-auto mt-4 max-w-xs'
        color='primary'
        disabled={ loading }
        onClick={ handleManageSubscription }
      >
        Manage Your Subscription
      </MBaseButton>
    </Fragment>
  );
}

interface TrialingRowProps {
  trialingUntil: Date | null;
}

function TrialingRow(props: TrialingRowProps): JSX.Element {
  if (!props.trialingUntil || isPast(props.trialingUntil)) {
    return null;
  }

  const trialEndDate = isThisYear(props.trialingUntil) ?
    format(props.trialingUntil, 'MMMM do') :
    format(props.trialingUntil, 'MMMM do, yyyy');

  return (
    <Fragment>
      <div className='flex justify-between py-4'>
        <MSpan>
          Subscription Status
        </MSpan>
        <MBadge className='bg-yellow-600'>
          <AccessTimeOutlined />
          Trialing Until { trialEndDate }
        </MBadge>
      </div>
      <MDivider />
      <div className='flex justify-between py-4'>
        <MSpan>
          You can upgrade to a paid subscription at the end of your trial.
        </MSpan>
      </div>
      <MDivider />
    </Fragment>
  );
}
