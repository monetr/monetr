import { formatDistanceToNow } from 'date-fns';

import Typography from '@monetr/interface/components/Typography';
import { useLink } from '@monetr/interface/hooks/useLink';

import styles from './PlaidLastUpdatedCard.module.scss';

interface PlaidLastUpdatedCardProps {
  linkId?: string;
}

export default function PlaidLastUpdatedCard(props: PlaidLastUpdatedCardProps): React.JSX.Element | null {
  const link = useLink(props?.linkId);

  if (!link?.data?.plaidLink) {
    return null;
  }

  const lastUpdateString = link?.data?.plaidLink?.lastSuccessfulUpdate
    ? formatDistanceToNow(link.data.plaidLink.lastSuccessfulUpdate)
    : 'Never';

  const lastAttemptString = link?.data?.plaidLink?.lastAttemptedUpdate
    ? formatDistanceToNow(link.data.plaidLink.lastAttemptedUpdate)
    : 'Never';

  return (
    <div className={styles.card}>
      <Typography color='subtle' ellipsis size='sm'>
        Last Updated: {lastUpdateString} ago
      </Typography>
      <Typography className={styles.attemptText} color='subtle' ellipsis size='sm'>
        Last Attempt: {lastAttemptString} ago
      </Typography>
    </div>
  );
}
