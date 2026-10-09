import { useEffect } from 'react';
import { LoaderCircle } from 'lucide-react';
import { useLocation, useSearch } from 'wouter';

import Logo from '@monetr/interface/assets/Logo';
import Typography from '@monetr/interface/components/Typography';
import { useAfterCheckout } from '@monetr/interface/hooks/useAfterCheckout';
import { useSnackbar } from '@monetr/notify';

import styles from './after.module.scss';

export default function AfterCheckoutPage(): React.JSX.Element {
  const search = useSearch();
  const [, navigate] = useLocation();
  const afterCheckout = useAfterCheckout();
  const { enqueueSnackbar } = useSnackbar();

  // As soon as the component mounts, call setup from checkout to get the subscription sorted out.
  useEffect(() => {
    const params = new URLSearchParams(search);
    const checkoutSessionId = params.get('session') ?? '';
    afterCheckout(checkoutSessionId)
      .then(result => {
        // If the user's subscription is now active then redirect them to the main view of the authenticated
        // application.
        if (result.isActive) {
          return navigate('/');
        }

        // Otherwise, dispaly the message from the result of the afterCheckout call.
        enqueueSnackbar(result?.message || 'Subscription is not active', {
          variant: 'error',
          disableWindowBlurListener: true,
        });
      })
      .catch(() =>
        enqueueSnackbar('Unable to determine your subscription state, please contact support@monetr.app', {
          variant: 'error',
          disableWindowBlurListener: true,
        }),
      );
  }, [search, afterCheckout, navigate, enqueueSnackbar]);

  return (
    <div className={styles.root}>
      <div className={styles.card}>
        <div className={styles.logoRow}>
          <Logo className={styles.logo} />
        </div>
        <div className={styles.row}>
          <Typography className={styles.message} size='xl'>
            Getting your account setup...
          </Typography>
        </div>
        <div className={styles.spinnerRow}>
          <LoaderCircle className={styles.spin} />
        </div>
      </div>
    </div>
  );
}
