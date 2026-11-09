import { formatDistanceToNow } from 'date-fns';

import Typography from '@monetr/interface/components/Typography';
import { useLink } from '@monetr/interface/hooks/useLink';

import styles from './PlaidLastUpdatedCard.module.scss';

interface SimpleFinLastUpdatedCardProps {
  linkId?: string;
}

export default function SimpleFinLastUpdatedCard(props: SimpleFinLastUpdatedCardProps): React.JSX.Element | null {
  const link = useLink(props?.linkId);

  if (!link?.data?.simpleFinLink) {
    return null;
  }

  const lastUpdateString = link.data.simpleFinLink.lastSuccessfulUpdate
    ? formatDistanceToNow(link.data.simpleFinLink.lastSuccessfulUpdate, { addSuffix: true })
    : 'Never';

  const lastAttemptString = link.data.simpleFinLink.lastAttemptedUpdate
    ? formatDistanceToNow(link.data.simpleFinLink.lastAttemptedUpdate, { addSuffix: true })
    : 'Never';

  return (
    <div className={styles.card}>
      <Typography color='subtle' ellipsis size='sm'>
        Last Updated: {lastUpdateString}
      </Typography>
      <Typography className={styles.attemptText} color='subtle' ellipsis size='sm'>
        Last Attempt: {lastAttemptString}
      </Typography>
    </div>
  );
}
