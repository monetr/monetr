import { formatDistanceToNow } from 'date-fns';
import { Check, RefreshCw, TriangleAlert } from 'lucide-react';

import Badge, { type BadgeProps } from '@monetr/interface/components/Badge';
import { Button } from '@monetr/interface/components/Button';
import Typography from '@monetr/interface/components/Typography';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { useTriggerManualLunchFlowSync } from '@monetr/interface/hooks/useTriggerManualLunchFlowSync';
import { useTriggerManualPlaidSync } from '@monetr/interface/hooks/useTriggerManualPlaidSync';
import type Link from '@monetr/interface/models/Link';
import { DateLength, formatDate } from '@monetr/interface/util/formatDate';

import styles from './LinkStatusCard.module.scss';

export interface LinkStatusCardProps {
  link: Link;
}

export default function LinkStatusCard(props: LinkStatusCardProps): React.JSX.Element | null {
  const { inTimezone } = useTimezone();
  const { data: locale } = useLocale();
  const triggerPlaidSync = useTriggerManualPlaidSync();
  const triggerLunchFlowSync = useTriggerManualLunchFlowSync();

  // Manual links don't sync with anything, so there is nothing to show here.
  const sync = props.link.plaidLink ?? props.link.lunchFlowLink;
  if (!sync) {
    return null;
  }

  let provider = 'Plaid';
  let status = 'Connected';
  let statusVariant: BadgeProps['variant'] = 'success';
  if (props.link.getIsLunchFlow()) {
    provider = 'Lunch Flow';
    status = 'Active';
  }
  if (props.link.getIsPendingExpiration()) {
    status = 'Expires Soon';
    statusVariant = 'warning';
  }
  if (props.link.getIsError()) {
    status = 'Needs Attention';
    statusVariant = 'destructive';
  }
  if (props.link.getIsRevoked()) {
    status = 'Disconnected';
    statusVariant = 'destructive';
  }
  if (props.link.getIsDeactivated()) {
    status = 'Deactivated';
    statusVariant = 'destructive';
  }

  let headline = 'Not Synced Yet';
  if (sync.lastSuccessfulUpdate) {
    headline = `Synced ${formatDistanceToNow(sync.lastSuccessfulUpdate, { addSuffix: true })}`;
  }

  const isError = props.link.getIsError();
  const expirationDate = props.link.plaidLink?.expirationDate;
  const canSync =
    (props.link.getIsPlaid() && !props.link.getIsRevoked()) ||
    (props.link.getIsLunchFlow() && !props.link.getIsDeactivated());

  function triggerSync() {
    if (props.link.getIsLunchFlow()) {
      triggerLunchFlowSync(props.link.linkId);
      return;
    }
    triggerPlaidSync(props.link.linkId);
  }

  return (
    <section className={styles.card}>
      <div className={styles.header}>
        <div className={styles.summary}>
          <div className={styles.icon} data-error={isError}>
            {isError && <TriangleAlert />}
            {!isError && <Check />}
          </div>
          <div className={styles.summaryText}>
            <div className={styles.providerRow}>
              <Typography className={styles.provider} color='subtle' size='xs' weight='semibold'>
                {provider}
              </Typography>
              <Badge size='xs' variant={statusVariant}>
                {status}
              </Badge>
            </div>
            <Typography color='emphasis' component='h3' size='lg' weight='semibold'>
              {headline}
            </Typography>
          </div>
        </div>
        {canSync && (
          <Button onClick={triggerSync} variant='secondary'>
            <RefreshCw />
            Sync Now
          </Button>
        )}
      </div>
      <dl className={styles.facts}>
        <div className={styles.fact}>
          <dt className={styles.factLabel}>Last Attempt</dt>
          <dd className={styles.factValue}>
            {sync.lastAttemptedUpdate ? formatDistanceToNow(sync.lastAttemptedUpdate, { addSuffix: true }) : 'Never'}
          </dd>
        </div>
        <div className={styles.fact}>
          <dt className={styles.factLabel}>Last Manual Sync</dt>
          <dd className={styles.factValue}>
            {sync.lastManualSync ? formatDistanceToNow(sync.lastManualSync, { addSuffix: true }) : 'Never'}
          </dd>
        </div>
        {props.link.getIsPendingExpiration() && expirationDate && (
          <div className={styles.fact}>
            <dt className={styles.factLabel}>Access Expires</dt>
            <dd className={styles.factValue} data-warning='true'>
              {locale ? formatDate(expirationDate, inTimezone, locale, DateLength.Medium) : ''}
            </dd>
          </div>
        )}
        {!props.link.getIsPendingExpiration() && (
          <div className={styles.fact}>
            <dt className={styles.factLabel}>Connected Since</dt>
            <dd className={styles.factValue}>
              {locale ? formatDate(props.link.createdAt, inTimezone, locale, DateLength.Medium) : ''}
            </dd>
          </div>
        )}
      </dl>
    </section>
  );
}
