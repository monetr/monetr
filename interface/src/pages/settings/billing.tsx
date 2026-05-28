import { useCallback, useState } from 'react';
import { format, isFuture, isThisYear } from 'date-fns';
import { Clock } from 'lucide-react';
import { useLocation } from 'wouter';

import type { ApiResponse } from '@monetr/interface/api/client';
import Badge from '@monetr/interface/components/Badge';
import { Button } from '@monetr/interface/components/Button';
import Divider from '@monetr/interface/components/Divider';
import Typography from '@monetr/interface/components/Typography';
import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';
import request from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

import styles from './billing.module.scss';

export default function SettingsBilling(): JSX.Element {
  const [pathname] = useLocation();
  const { enqueueSnackbar } = useSnackbar();
  const [loading, setLoading] = useState(false);
  const { data: auth } = useAuthentication();
  const handleManageSubscription = useCallback(async () => {
    setLoading(true);
    let promise: Promise<ApiResponse<{ url: string }>>;
    if (!auth?.hasSubscription) {
      promise = request({
        method: 'POST',
        url: '/api/billing/create_checkout',
        data: {
          // If the user backs out of the stripe checkout then return them to the current URL.
          cancelPath: pathname,
        },
      });
    } else {
      // If the customer has a subscription then we want to just manage it. This will allow a customer to fix a
      // subscription for a card that has failed payment or something similar.
      promise = request({ method: 'GET', url: '/api/billing/portal' });
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
  }, [enqueueSnackbar, auth, pathname]);

  const manageSubscriptionText = auth?.hasSubscription ? 'Manage Your Subscription' : 'Subscribe Early';

  return (
    <div className={styles.root}>
      <Typography className={styles.heading} color='emphasis' size='2xl' weight='bold'>
        Billing
      </Typography>
      <Divider />

      <div className={styles.statusRow}>
        <Typography size='inherit'>Subscription Status</Typography>
        <SubscriptionStatusBadge />
      </div>
      <Divider />

      <Button
        className={styles.subscribeButton}
        data-testid='billing-subscribe'
        disabled={loading}
        onClick={handleManageSubscription}
        variant='primary'
      >
        {manageSubscriptionText}
      </Button>
    </div>
  );
}

function SubscriptionStatusBadge(): JSX.Element {
  const { data: auth } = useAuthentication();

  // If they have a subscription and it is active then show active.
  if (auth?.hasSubscription && auth?.isActive) {
    return (
      <Badge data-testid='billing-subscription-active' variant='success'>
        Active
      </Badge>
    );
  }

  // If they have a trial end date that is in the future then they are trialing.
  if (auth?.trialingUntil && isFuture(auth?.trialingUntil)) {
    const trialEndDate = isThisYear(auth?.trialingUntil)
      ? format(auth?.trialingUntil, 'MMMM do')
      : format(auth?.trialingUntil, 'MMMM do, yyyy');

    return (
      <Badge data-testid='billing-subscription-trialing' variant='warning'>
        <Clock />
        Trialing Until {trialEndDate}
      </Badge>
    );
  }

  // Anything else is considered expired.
  return (
    <Badge data-testid='billing-subscription-expired' variant='destructive'>
      Expired
    </Badge>
  );
}
