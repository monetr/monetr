import { format } from 'date-fns';
import { Repeat } from 'lucide-react';
import { rrulestr } from 'rrule';

import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import RecurringMemo from '@monetr/interface/components/recurring/RecurringMemo';
import { getConfidenceLabel } from '@monetr/interface/components/transactions/TransactionRecurringCard';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { useTransactionCluster } from '@monetr/interface/hooks/useTransactionCluster';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import { AmountType } from '@monetr/interface/util/amounts';
import { DateLength, formatDate, formatRelativeDate } from '@monetr/interface/util/formatDate';

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

  let eyebrow = 'Recurring Deposit';
  let lastLabel = 'Last Deposit';
  if (recurring.direction === 'debit') {
    eyebrow = 'Recurring Charge';
    lastLabel = 'Last Charge';
  }

  let status = 'Active';
  if (recurring.ended) {
    status = 'Ended';
  }

  // The next three times we expect it. The server already worked out the first one so start from that.
  const rule = rrulestr(recurring.ruleset);
  const comingUp = [recurring.next];
  let last = recurring.next;
  for (let i = 0; i < 2; i++) {
    const after = rule.after(last, false);
    if (!after) {
      break;
    }
    comingUp.push(after);
    last = after;
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
            <span className={styles.eyebrow}>{eyebrow}</span>
            <span className={styles.statusBadge} data-ended={String(recurring.ended)}>
              {status}
            </span>
          </div>
          <span className={styles.title}>
            {locale.formatAmount(Math.abs(recurring.lastAmount), AmountType.Stored)} {rule.toText()}
          </span>
          {cluster?.originalMemo && (
            <div className={styles.showsUpAs}>
              <span className={styles.showsUpAsLabel}>Shows Up As</span>
              <RecurringMemo className={styles.memo} memo={cluster.originalMemo} />
            </div>
          )}
        </div>
      </div>

      <div className={styles.stats}>
        {recurring.ended && (
          <div className={styles.stat}>
            <span className={styles.statLabel}>Last Seen</span>
            <span className={styles.statValue}>
              {formatDate(recurring.last, inTimezone, dateLocale, DateLength.Medium)}
            </span>
          </div>
        )}
        {!recurring.ended && (
          <div className={styles.stat}>
            <span className={styles.statLabel}>Next Expected</span>
            <span className={styles.statValue}>
              {formatDate(recurring.next, inTimezone, dateLocale, DateLength.Medium)}
            </span>
            <span className={styles.statDetail}>{formatRelativeDate(recurring.next, inTimezone, dateLocale)}</span>
          </div>
        )}
        <div className={styles.stat}>
          <span className={styles.statLabel}>{lastLabel}</span>
          <span className={styles.statValue}>
            {formatDate(recurring.last, inTimezone, dateLocale, DateLength.Medium)}
          </span>
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
          <span className={styles.statLabel}>Coming Up</span>
          <div className={styles.comingUpDates}>
            {comingUp.map((item, index) => (
              <span className={styles.comingUpDate} data-next={String(index === 0)} key={item.toISOString()}>
                {formatDate(item, inTimezone, dateLocale, DateLength.Medium)}
              </span>
            ))}
          </div>
        </div>
      )}
    </section>
  );
}
