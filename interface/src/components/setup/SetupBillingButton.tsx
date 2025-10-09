import React, { useCallback, useState } from 'react';
import { CreditCard } from '@mui/icons-material';
import axios from 'axios';
import { useSnackbar } from 'notistack';

import { MBaseButton } from '@monetr/interface/components/MButton';
import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';


/**
 * The SetupBillingButton should only be used on the setup page, it is intended to be a way to manage your billing
 * settings if you do not have an active link for some reason.
 */
export default function SetupBillingButton(): JSX.Element {
  const { data: { hasSubscription } } = useAuthentication();
  const { enqueueSnackbar } = useSnackbar();
  const [loading, setLoading] = useState(false);

  const handleManageSubscription = useCallback(() => {
    setLoading(true);
    // If the customer has a subscription then we want to just manage it. This will allow a customer to fix a
    // subscription for a card that has failed payment or something similar.
    axios.get('/api/billing/portal')
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
    <MBaseButton
      className='max-w-xs'
      color='secondary'
      disabled={ loading }
      onClick={ handleManageSubscription }
    >
      <CreditCard className='mr-2' />
      Manage Your Subscription
    </MBaseButton>
  );
}
