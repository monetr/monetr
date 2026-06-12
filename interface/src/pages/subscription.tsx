import { LoaderCircle } from 'lucide-react';
import { useLocation } from 'wouter';

import Logo from '@monetr/interface/assets/Logo';
import Typography from '@monetr/interface/components/Typography';
import useMountEffect from '@monetr/interface/hooks/useMountEffect';
import request from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

import styles from './subscription.module.scss';

// SubscriptionPage is just used to redirect the user to the stripe billing portal. Upon mounting, it will make an API
// call to start a billing portal session, and once it gets a response it will redirect the user there.
export default function SubscriptionPage(): React.JSX.Element {
  const { enqueueSnackbar } = useSnackbar();
  const [, navigate] = useLocation();

  useMountEffect(() => {
    request<{ url: string }>({ method: 'GET', url: '/api/billing/portal' })
      .then(result => window.location.assign(result.data.url))
      .catch(error => {
        enqueueSnackbar(error?.response?.data?.error || 'Failed to navigate to billing portal.', {
          variant: 'error',
          disableWindowBlurListener: true,
        });
        navigate('/');
      });
  });

  return (
    <div className={styles.root}>
      <div className={styles.card}>
        <div className={styles.logoRow}>
          <Logo className={styles.logo} />
        </div>
        <div className={styles.row}>
          <Typography className={styles.message} size='xl'>
            Loading the billing portal...
          </Typography>
        </div>
        <div className={styles.spinnerRow}>
          <LoaderCircle className={styles.spin} />
        </div>
      </div>
    </div>
  );
}
