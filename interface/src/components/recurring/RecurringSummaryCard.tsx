import { format, isThisYear } from 'date-fns';
import { Repeat } from 'lucide-react';
import { rrulestr } from 'rrule';

import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import { getConfidenceLabel } from '@monetr/interface/components/transactions/TransactionRecurringCard';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { useTransactionCluster } from '@monetr/interface/hooks/useTransactionCluster';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import { AmountType } from '@monetr/interface/util/amounts';
import { formatRelativeDate } from '@monetr/interface/util/formatDate';

import styles from './RecurringSummaryCard.module.scss';

export interface RecurringSummaryCardProps {
  recurring: TransactionRecurring;
}

export default function RecurringSummaryCard({ recurring }: RecurringSummaryCardProps): React.JSX.Element | null {
  const { data: locale } = useLocaleCurrency();
  const { data: dateLocale } = useLocale();
  const { inTimezone } = useTimezone();
  const { data: cluster } = useTransactionCluster(recurring.transactionClusterId);

  if (!locale || !dateLocale) {
    return null;
  }

  const isDebit = recurring.direction === 'debit';
  const memo = cluster?.originalMemo;

  function formatShortDate(date: Date): string {
    if (isThisYear(date)) {
      return format(inTimezone(date), 'MMM d');
    }

    return format(inTimezone(date), 'MMM d, yyyy');
  }

  // The next three times we expect it. The server already worked out the first one so start from that.
  const rule = rrulestr(recurring.ruleset);
  const comingUp = [recurring.next];
  for (let i = 0; i < 2; i++) {
    const after = rule.after(comingUp[comingUp.length - 1] ?? recurring.next, false);
    if (!after) {
      break;
    }
    comingUp.push(after);
  }

  return (
    <section className={styles.root}>
      <div className={styles.header}>
        <div className={styles.icon}>
          <MerchantIcon name={cluster?.name} />
          <span className={styles.iconBadge} title='Recurring'>
            <Repeat />
          </span>
        </div>
        <div className={styles.headerText}>
          <div className={styles.eyebrowRow}>
            <span className={styles.eyebrow}>{isDebit ? 'Recurring charge' : 'Recurring deposit'}</span>
            <span className={styles.statusBadge} data-ended={String(recurring.ended)}>
              {recurring.ended ? 'Ended' : 'Active'}
            </span>
          </div>
          <span className={styles.title}>
            {locale.formatAmount(Math.abs(recurring.lastAmount), AmountType.Stored)} {rule.toText()}
          </span>
          {memo && (
            <div className={styles.showsUpAs}>
              <span className={styles.showsUpAsLabel}>Shows up as</span>
              <span className={styles.memo} title={memo}>
                {memo}
              </span>
            </div>
          )}
        </div>
      </div>

      <div className={styles.stats}>
        {recurring.ended && (
          <div className={styles.stat}>
            <span className={styles.statLabel}>Last seen</span>
            <span className={styles.statValue}>{formatShortDate(recurring.last)}</span>
          </div>
        )}
        {!recurring.ended && (
          <div className={styles.stat}>
            <span className={styles.statLabel}>Next expected</span>
            <span className={styles.statValue}>{formatShortDate(recurring.next)}</span>
            <span className={styles.statDetail}>{formatRelativeDate(recurring.next, inTimezone, dateLocale)}</span>
          </div>
        )}
        <div className={styles.stat}>
          <span className={styles.statLabel}>{isDebit ? 'Last charge' : 'Last deposit'}</span>
          <span className={styles.statValue}>{formatShortDate(recurring.last)}</span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statLabel}>Since</span>
          <span className={styles.statValue}>{format(inTimezone(recurring.first), 'MMM yyyy')}</span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statLabel}>Confidence</span>
          <span className={styles.statValue} title={`${Math.round(recurring.confidence * 100)}%`}>
            {getConfidenceLabel(recurring.confidence)}
          </span>
        </div>
      </div>

      {!recurring.ended && (
        <div className={styles.comingUp}>
          <span className={styles.statLabel}>Coming up</span>
          <div className={styles.comingUpDates}>
            {comingUp.map((date, index) => (
              <span className={styles.comingUpDate} data-next={String(index === 0)} key={date.toISOString()}>
                {formatShortDate(date)}
              </span>
            ))}
          </div>
        </div>
      )}
    </section>
  );
}
