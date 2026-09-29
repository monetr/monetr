import { formatDistanceToNow } from 'date-fns';

import Typography from '@monetr/interface/components/Typography';
import { useLink } from '@monetr/interface/hooks/useLink';

import styles from './PlaidLastUpdatedCard.module.scss';

interface LunchFlowLastUpdatedCardProps {
  linkId?: string;
}

export default function LunchFlowLastUpdatedCard(props: LunchFlowLastUpdatedCardProps): React.JSX.Element | null {
  const link = useLink(props?.linkId);

  if (!link?.data?.lunchFlowLink) {
    return null;
  }

  const lastUpdateString = link.data.lunchFlowLink.lastSuccessfulUpdate
    ? formatDistanceToNow(link.data.lunchFlowLink.lastSuccessfulUpdate, { addSuffix: true })
    : 'Never';

  const lastAttemptString = link.data.lunchFlowLink.lastAttemptedUpdate
    ? formatDistanceToNow(link.data.lunchFlowLink.lastAttemptedUpdate, { addSuffix: true })
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
