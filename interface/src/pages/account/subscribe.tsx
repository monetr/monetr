import { useCallback, useState } from 'react';
import { useSnackbar } from 'notistack';

import { Button } from '@monetr/interface/components/Button';
import Flex from '@monetr/interface/components/Flex';
import { layoutVariants } from '@monetr/interface/components/Layout';
import MLogo from '@monetr/interface/components/MLogo';
import TextLink from '@monetr/interface/components/TextLink';
import Typography from '@monetr/interface/components/Typography';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';
import request from '@monetr/interface/util/request';

import styles from './subscribe.module.scss';

export default function SubscribePage(): JSX.Element {
  const { enqueueSnackbar } = useSnackbar();
  const {
    data: { initialPlan },
  } = useAppConfiguration();
  const {
    data: { hasSubscription, activeUntil },
  } = useAuthentication();

  const [loading, setLoading] = useState(false);
  const handleContinue = useCallback(() => {
    setLoading(true);
    let promise: Promise<{ data: { url: string } }>;
    if (initialPlan && !hasSubscription) {
      promise = request().post('/billing/create_checkout', {
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
      <div className={styles.subscribePageRoot}>
        <Flex align='center' flex='grow' justify='between' orientation='column'>
          <div />
          <Flex align='center' className={layoutVariants({ maxWidth: 'medium' })} orientation='column'>
            <MLogo className={layoutVariants({ size: 'logo' })} />
            <Typography align='center' size='2xl' weight='semibold'>
              Your subscription is no longer active
            </Typography>
            <Typography align='center' size='lg'>
              Thank you for having subscribed to monetr before! If you'd like to continue using monetr you will have to
              resubscribe below. Click continue to proceed to our billing portal.
            </Typography>
            <Button disabled={loading} onClick={handleContinue} variant='primary'>
              Continue
            </Button>
          </Flex>
          {!loading && (
            <Flex align='center' gap='sm' justify='center'>
              <Typography color='subtle' size='sm'>
                Not ready to continue?
              </Typography>
              <TextLink size='sm' to='/logout'>
                Logout for now
              </TextLink>
            </Flex>
          )}
        </Flex>
      </div>
    );
  }

  return (
    <div className={styles.subscribePageRoot}>
      <Flex align='center' flex='grow' justify='between' orientation='column'>
        <div />
        <Flex align='center' className={layoutVariants({ maxWidth: 'medium' })} orientation='column'>
          <MLogo className={layoutVariants({ size: 'logo' })} />
          <Typography align='center' size='2xl' weight='semibold'>
            Your free trial has ended
          </Typography>
          <Typography align='center' size='lg'>
            Thank you for trying out monetr! We hope that you found our budgeting tools useful during your trial. If
            you'd like to continue using monetr you can easily subscribe below.
          </Typography>
          <Button disabled={loading} onClick={handleContinue} variant='primary'>
            Continue
          </Button>
        </Flex>
        {loading && <div />}
        {!loading && (
          <Flex align='center' gap='sm' justify='center'>
            <Typography color='subtle' size='sm'>
              Not ready to continue?
            </Typography>
            <TextLink size='sm' to='/logout'>
              Logout for now
            </TextLink>
          </Flex>
        )}
      </Flex>
    </div>
  );
}
