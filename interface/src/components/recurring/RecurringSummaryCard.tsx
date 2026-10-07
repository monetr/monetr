import { Fragment } from 'react';
import { addDays } from 'date-fns';
import { Check, CircleAlert, Repeat } from 'lucide-react';
import { rrulestr } from 'rrule';

import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import RecurringMemo from '@monetr/interface/components/recurring/RecurringMemo';
import { Tooltip, TooltipContent, TooltipTrigger } from '@monetr/interface/components/Tooltip';
import Typography from '@monetr/interface/components/Typography';
import { getConfidenceLabel } from '@monetr/interface/components/transactions/TransactionRecurringCard';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useSpending } from '@monetr/interface/hooks/useSpending';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { useTransactionCluster } from '@monetr/interface/hooks/useTransactionCluster';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import { AmountType } from '@monetr/interface/util/amounts';
import { DateLength, formatDate, formatRelativeDate } from '@monetr/interface/util/formatDate';

import styles from './RecurringSummaryCard.module.scss';

export interface RecurringSummaryCardProps {
  recurring: TransactionRecurring;
}

export default function RecurringSummaryCard(props: RecurringSummaryCardProps): React.JSX.Element | null {
  const { data: locale } = useLocaleCurrency();
  const { data: dateLocale } = useLocale();
  const { inTimezone } = useTimezone();
  const { data: cluster } = useTransactionCluster(props.recurring.transactionClusterId);
  const { data: spending } = useSpending(props.recurring.spendingId);

  if (!locale || !dateLocale) {
    return null;
  }

  let nextLabel = 'Next Deposit';
  if (props.recurring.direction === 'debit') {
    nextLabel = 'Next Charge';
  }

  let status = 'Active';
  if (props.recurring.ended) {
    status = 'Ended';
  }

  // The couple of times we expect it after the next one. The server already worked out the next one so start from that,
  // but look from the day after. Next is midnight in the account's timezone and the rule can be an hour off from that
  // when daylight savings changed since it started, so it would just give us next again.
  const rule = rrulestr(props.recurring.ruleset);
  const later: Array<Date> = [];
  let last = props.recurring.next;
  for (let i = 0; i < 2; i++) {
    const after = rule.after(addDays(last, 1), true);
    if (!after) {
      break;
    }
    later.push(after);
    last = after;
  }

  const amount = locale.formatAmount(Math.abs(props.recurring.lastAmount), AmountType.Stored);
  // If the expense isn't behind then it'll have enough by the time the next charge comes in, even if it doesn't yet
  const covered = !spending?.isBehind;

  return (
    <Fragment>
      <section className={styles.hero}>
        <div className={styles.icon}>
          <MerchantIcon name={cluster?.name} size='large' />
          <span className={styles.iconBadge} title='Recurring'>
            <Repeat />
          </span>
        </div>
        <div className={styles.heroText}>
          <Typography color='emphasis' component='h1' ellipsis size='3xl' weight='bold'>
            {cluster?.name}
          </Typography>
          {cluster?.originalMemo && <RecurringMemo className={styles.memo} memo={cluster.originalMemo} />}
          <div className={styles.scheduleRow}>
            <span className={styles.schedule}>
              <strong>{amount}</strong> {rule.toText()}
            </span>
            <span className={styles.statusBadge} data-ended={String(props.recurring.ended)}>
              {status}
            </span>
            <Tooltip delayDuration={100}>
              <TooltipTrigger asChild>
                <span className={styles.confidenceBadge}>{getConfidenceLabel(props.recurring.confidence)}</span>
              </TooltipTrigger>
              <TooltipContent side='top'>Confidence {Math.round(props.recurring.confidence * 100)}%</TooltipContent>
            </Tooltip>
          </div>
        </div>
      </section>

      {!props.recurring.ended && (
        <section className={styles.next}>
          <div className={styles.nextText}>
            <span className={styles.nextLabel}>{nextLabel}</span>
            <span className={styles.nextDate}>
              {formatDate(props.recurring.next, inTimezone, dateLocale, DateLength.Long)}{' '}
              <span className={styles.nextRelative}>
                {formatRelativeDate(props.recurring.next, inTimezone, dateLocale)}
              </span>
            </span>
            {later.length > 0 && (
              <span className={styles.nextLater}>
                Then {later.map(item => formatDate(item, inTimezone, dateLocale, DateLength.Medium)).join(' and ')}
              </span>
            )}
          </div>
          {spending && (
            <div className={styles.coverage} data-covered={String(covered)}>
              {covered && <Check />}
              {!covered && <CircleAlert />}
              <div className={styles.coverageText}>
                <span className={styles.coverageTitle}>
                  {covered && 'Covered'}
                  {!covered && 'Not Covered'}
                </span>
                <span className={styles.coverageDetail}>
                  {covered && `${spending.name} will have ${amount} ready`}
                  {!covered &&
                    `${spending.name} only has ${locale.formatAmount(spending.currentAmount, AmountType.Stored)} of ${amount}`}
                </span>
              </div>
            </div>
          )}
        </section>
      )}
    </Fragment>
  );
}
