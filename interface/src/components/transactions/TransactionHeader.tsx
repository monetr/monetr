import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import { Skeleton } from '@monetr/interface/components/Skeleton';
import Typography from '@monetr/interface/components/Typography';
import TransactionAmount from '@monetr/interface/components/transactions/TransactionAmount';
import TransactionRecurringIndicator from '@monetr/interface/components/transactions/TransactionRecurringIndicator';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import type Transaction from '@monetr/interface/models/Transaction';
import { DateLength, formatDate } from '@monetr/interface/util/formatDate';

import styles from './TransactionHeader.module.scss';

export interface TransactionHeaderProps {
  // Will show a placeholder while the transaction is still loading
  transaction?: Transaction;
}

export default function TransactionHeader(props: TransactionHeaderProps): React.JSX.Element {
  const { data: dateLocale } = useLocale();
  const { inTimezone } = useTimezone();

  if (!props.transaction || !dateLocale) {
    return (
      <section className={styles.hero}>
        <MerchantIcon size='large' />
        <div className={styles.heroText}>
          <Skeleton className={styles.titlePlaceholder} />
          <Skeleton className={styles.detailsPlaceholder} />
        </div>
      </section>
    );
  }

  return (
    <section className={styles.hero}>
      <div className={styles.icon}>
        <MerchantIcon name={props.transaction.name ?? undefined} size='large' />
        {props.transaction.transactionRecurringId && (
          <TransactionRecurringIndicator transactionRecurringId={props.transaction.transactionRecurringId} />
        )}
      </div>
      <div className={styles.heroText}>
        <Typography color='emphasis' component='h1' ellipsis='mobileScroll' size='3xl' weight='bold'>
          {props.transaction.name}
        </Typography>
        {props.transaction.originalName && (
          <Typography
            className={styles.originalName}
            color='subtle'
            ellipsis='mobileScroll'
            size='xs'
            title={props.transaction.originalName}
          >
            {props.transaction.originalName}
          </Typography>
        )}
        <div className={styles.detailsRow}>
          <span className={styles.details}>
            <TransactionAmount transaction={props.transaction} />{' '}
            {formatDate(props.transaction.date, inTimezone, dateLocale, DateLength.Long)}
          </span>
          {props.transaction.isPending && (
            <span className={styles.pendingBadge} data-testid='transaction-details-pending-badge'>
              Pending
            </span>
          )}
        </div>
      </div>
    </section>
  );
}
