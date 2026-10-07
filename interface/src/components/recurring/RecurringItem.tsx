import { useState } from 'react';
import { differenceInCalendarDays, format, isThisYear, startOfToday } from 'date-fns';
import { CalendarSync, ChevronDown, ChevronUp, Receipt } from 'lucide-react';
import { rrulestr } from 'rrule';
import { Link } from 'wouter';

import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import RecurringItemCharges from '@monetr/interface/components/recurring/RecurringItemCharges';
import Typography from '@monetr/interface/components/Typography';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useSpending } from '@monetr/interface/hooks/useSpending';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { showNewExpenseModal } from '@monetr/interface/modals/NewExpenseModal';
import { showNewFundingModal } from '@monetr/interface/modals/NewFundingModal';
import type Spending from '@monetr/interface/models/Spending';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import { TransactionRecurringWindow } from '@monetr/interface/models/TransactionRecurring';
import { AmountType } from '@monetr/interface/util/amounts';
import capitalize from '@monetr/interface/util/capitalize';
import { formatRelativeDate } from '@monetr/interface/util/formatDate';

import styles from './RecurringItem.module.scss';

export interface RecurringItemProps {
  recurring: TransactionRecurring;
}

// Short version of the schedule, the full ruleset text is on the details page.
const cadences: Record<TransactionRecurringWindow, string> = {
  [TransactionRecurringWindow.FirstAndFifteenth]: 'Twice a month',
  [TransactionRecurringWindow.FifteenthAndLast]: 'Twice a month',
  [TransactionRecurringWindow.Weekly]: 'Weekly',
  [TransactionRecurringWindow.BiWeekly]: 'Every 2 weeks',
  [TransactionRecurringWindow.Monthly]: 'Monthly',
  [TransactionRecurringWindow.BiMonthly]: 'Every 2 months',
  [TransactionRecurringWindow.Quarterly]: 'Quarterly',
  [TransactionRecurringWindow.SemiYearly]: 'Every 6 months',
  [TransactionRecurringWindow.Yearly]: 'Yearly',
};

// How much of the next charge the expense won't cover. Only counts if the expense is actually behind, if its going to
// catch up before the charge comes in then its not really short.
function getShortfall(recurring: TransactionRecurring, spending: Spending | undefined): number {
  if (recurring.ended || !spending?.isBehind) {
    return 0;
  }

  return Math.max(0, Math.abs(recurring.lastAmount) - spending.currentAmount);
}

export default function RecurringItem({ recurring }: RecurringItemProps): React.JSX.Element | null {
  const { data: locale } = useLocaleCurrency();
  const { data: dateLocale } = useLocale();
  const { inTimezone } = useTimezone();
  // The list doesn't include the spending object, so look it up. This way it comes from the same cache as everything
  // else that shows spending.
  const { data: spending } = useSpending(recurring.spendingId);
  const [expanded, setExpanded] = useState(false);

  if (!locale || !dateLocale) {
    return null;
  }

  const name = recurring.transactionCluster?.name ?? 'Recurring';
  const memo = recurring.transactionCluster?.originalMemo;
  const detailsPath = `/bank/${recurring.bankAccountId}/recurring/${recurring.transactionRecurringId}/details`;
  const cadence = cadences[recurring.window] ?? capitalize(rrulestr(recurring.ruleset).toText());
  const isDebit = recurring.direction === 'debit';
  const isLinked = isDebit ? Boolean(recurring.spendingId) : Boolean(recurring.fundingSchedule);
  const linkedName = isDebit ? spending?.name : recurring.fundingSchedule?.name;
  const LinkedIcon = isDebit ? Receipt : CalendarSync;
  const shortfall = getShortfall(recurring, spending);
  const daysLate = differenceInCalendarDays(startOfToday({ in: inTimezone }), inTimezone(recurring.next));
  const isLate = !recurring.ended && daysLate > 0;

  // Ended ones show when they were last seen, everything else shows when its expected next.
  const date = recurring.ended ? recurring.last : recurring.next;
  const dateString = isThisYear(date) ? format(inTimezone(date), 'MMM d') : format(inTimezone(date), 'MMM d, yyyy');

  let dateDetail = formatRelativeDate(recurring.next, inTimezone, dateLocale);
  if (recurring.ended) {
    dateDetail = 'last seen';
  } else if (isLate) {
    dateDetail = `was due ${dateString}`;
  }

  // A price change shows up as two different amounts, so only call it variable once there are more than that.
  const amounts = Object.keys(recurring.amounts).map(amount => Math.abs(Number(amount)));
  const isVariable = amounts.length >= 3;
  let amountString = locale.formatAmount(Math.abs(recurring.lastAmount), AmountType.Stored);
  if (isVariable) {
    amountString = `~${amountString}`;
  }

  function createLink() {
    if (isDebit) {
      showNewExpenseModal({ recurring });
    } else {
      showNewFundingModal({ recurring });
    }
  }

  return (
    <li className={styles.root} data-testid={recurring.transactionRecurringId}>
      <div className={styles.container} data-expanded={String(expanded)}>
        <div className={styles.inner}>
          <Link aria-label={name} className={styles.link} to={detailsPath} />
          <MerchantIcon name={name} />
          <div className={styles.nameColumn}>
            <div className={styles.nameRow}>
              <Typography className={styles.name} color='emphasis' ellipsis weight='semibold'>
                {name}
              </Typography>
              {/* This block only shows on mobile screens */}
              <span className={styles.mobileCadence}>&middot; {cadence}</span>
            </div>
            {memo && (
              <span className={styles.memo} title={memo}>
                {memo}
              </span>
            )}
            {/* This block only shows on mobile screens */}
            <span className={styles.mobileStatus} data-warning={String(isDebit && !isLinked && !recurring.ended)}>
              {isLinked && <LinkedIcon />}
              {isLinked && <span className={styles.mobileStatusName}>{linkedName}</span>}
              {shortfall > 0 && (
                <span className={styles.short}>&middot; short {locale.formatAmount(shortfall, AmountType.Stored)}</span>
              )}
              {!isLinked && (isDebit ? 'Not budgeted' : 'Not funding anything')}
            </span>
          </div>

          {/* This block only shows on desktop screens */}
          <div className={styles.desktopDate}>
            {isLate && (
              <span className={styles.lateBadge}>{daysLate === 1 ? '1 day late' : `${daysLate} days late`}</span>
            )}
            {!isLate && <span className={styles.dateValue}>{dateString}</span>}
            <span className={styles.dateDetail}>
              {cadence} &middot; {dateDetail}
            </span>
          </div>

          {/* This block only shows on desktop screens */}
          <div className={styles.desktopBudget}>
            {isLinked && (
              <span className={styles.linkedIcon}>
                <LinkedIcon />
              </span>
            )}
            {isLinked && <span className={styles.linkedName}>{linkedName}</span>}
            {shortfall > 0 && (
              <span className={styles.short}>short {locale.formatAmount(shortfall, AmountType.Stored)}</span>
            )}
            {!isLinked && recurring.ended && (
              <span className={styles.notLinked}>{isDebit ? 'Not budgeted' : 'Not funding anything'}</span>
            )}
            {!isLinked && !recurring.ended && (
              <button className={styles.createButton} onClick={createLink} type='button'>
                {isDebit ? 'Create expense' : 'Create funding schedule'}
              </button>
            )}
          </div>

          <div className={styles.amountColumn}>
            <span className={styles.amount}>{amountString}</span>
            {isVariable && (
              <span className={styles.desktopRange}>
                {locale.formatAmount(Math.min(...amounts), AmountType.Stored)} to{' '}
                {locale.formatAmount(Math.max(...amounts), AmountType.Stored)}
              </span>
            )}
            {/* On mobile there is no date column so the date goes under the amount instead */}
            <span className={styles.mobileDate}>{dateString}</span>
          </div>

          <button
            aria-expanded={expanded}
            aria-label={`${expanded ? 'Hide' : 'Show'} recent ${name} charges`}
            className={styles.toggle}
            onClick={() => setExpanded(!expanded)}
            type='button'
          >
            {expanded ? <ChevronUp /> : <ChevronDown />}
          </button>
        </div>
        {expanded && <RecurringItemCharges detailsPath={detailsPath} name={name} recurring={recurring} />}
      </div>
    </li>
  );
}
