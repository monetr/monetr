import { useState } from 'react';
import { differenceInCalendarDays, startOfToday } from 'date-fns';
import { CalendarSync, ChevronDown, ChevronUp, Receipt } from 'lucide-react';
import { Link } from 'wouter';

import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import RecurringItemCharges from '@monetr/interface/components/recurring/RecurringItemCharges';
import RecurringMemo from '@monetr/interface/components/recurring/RecurringMemo';
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
import { DateLength, formatDate, formatRelativeDate } from '@monetr/interface/util/formatDate';

import styles from './RecurringItem.module.scss';

export interface RecurringItemProps {
  recurring: TransactionRecurring;
}

// Short version of the schedule, the full ruleset text is on the details page
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
  // The list doesn't include the spending object so look it up, this way it comes from the same cache as everything
  // else that shows spending
  const { data: spending } = useSpending(recurring.spendingId);
  const [expanded, setExpanded] = useState(false);

  if (!locale || !dateLocale) {
    return null;
  }

  const name = recurring.transactionCluster?.name ?? 'Recurring';
  const memo = recurring.transactionCluster?.originalMemo;
  const detailsPath = `/bank/${recurring.bankAccountId}/recurring/${recurring.transactionRecurringId}/details`;
  const cadence = cadences[recurring.window];
  const isDebit = recurring.direction === 'debit';
  const shortfall = getShortfall(recurring, spending);
  const today = startOfToday({
    in: inTimezone,
  });
  const daysLate = differenceInCalendarDays(today, inTimezone(recurring.next));
  const isLate = !recurring.ended && daysLate > 0;

  let isLinked = Boolean(recurring.fundingSchedule);
  let linkedName = recurring.fundingSchedule?.name;
  let notLinked = 'Not funding anything';
  let LinkedIcon = CalendarSync;
  if (isDebit) {
    isLinked = Boolean(recurring.spendingId);
    linkedName = spending?.name;
    notLinked = 'Not budgeted';
    LinkedIcon = Receipt;
  }

  // Ended ones show when they were last seen, everything else shows when its expected next
  let dateString = formatDate(recurring.next, inTimezone, dateLocale, DateLength.Medium);
  let dateDetail = formatRelativeDate(recurring.next, inTimezone, dateLocale);
  if (recurring.ended) {
    dateString = formatDate(recurring.last, inTimezone, dateLocale, DateLength.Medium);
    dateDetail = 'last seen';
  } else if (isLate) {
    dateDetail = `was due ${dateString}`;
  }

  let lateString = `${daysLate} days late`;
  if (daysLate === 1) {
    lateString = '1 day late';
  }

  // A price change shows up as two different amounts, so only call it variable once there are more than that
  const amounts = Object.keys(recurring.amounts).map(item => Math.abs(Number(item)));
  const isVariable = amounts.length >= 3;
  let amountString = locale.formatAmount(Math.abs(recurring.lastAmount), AmountType.Stored);
  if (isVariable) {
    amountString = `~${amountString}`;
  }

  let toggleLabel = `Show recent ${name} charges`;
  if (expanded) {
    toggleLabel = `Hide recent ${name} charges`;
  }

  function createLink() {
    if (isDebit) {
      showNewExpenseModal({
        recurring,
      });
    } else {
      showNewFundingModal({
        recurring,
      });
    }
  }

  return (
    <li className={styles.root}>
      <div className={styles.container} data-expanded={String(expanded)}>
        <div className={styles.inner}>
          <Link aria-label={name} className={styles.link} to={detailsPath} />
          <MerchantIcon name={name} />
          <div className={styles.nameColumn}>
            <div className={styles.nameRow}>
              <Typography
                className={styles.name}
                color='emphasis'
                data-testid='recurring-item-name'
                ellipsis
                weight='semibold'
              >
                {name}
              </Typography>
              {/* This block only shows on mobile screens */}
              <span className={styles.mobileCadence}>&middot; {cadence}</span>
            </div>
            {memo && <RecurringMemo memo={memo} />}
            {/* This block only shows on mobile screens */}
            <span className={styles.mobileStatus} data-warning={String(isDebit && !isLinked && !recurring.ended)}>
              {isLinked && <LinkedIcon />}
              {isLinked && <span className={styles.mobileStatusName}>{linkedName}</span>}
              {shortfall > 0 && (
                <span className={styles.short}>&middot; short {locale.formatAmount(shortfall, AmountType.Stored)}</span>
              )}
              {!isLinked && notLinked}
            </span>
          </div>

          {/* This block only shows on desktop screens */}
          <div className={styles.desktopDate}>
            {isLate && <span className={styles.lateBadge}>{lateString}</span>}
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
            {isLinked && (
              <span className={styles.linkedName} data-testid='recurring-item-linked'>
                {linkedName}
              </span>
            )}
            {shortfall > 0 && (
              <span className={styles.short} data-testid='recurring-item-short'>
                short {locale.formatAmount(shortfall, AmountType.Stored)}
              </span>
            )}
            {!isLinked && recurring.ended && <span className={styles.notLinked}>{notLinked}</span>}
            {!isLinked && !recurring.ended && (
              <button
                className={styles.createButton}
                data-testid='recurring-item-create'
                onClick={createLink}
                type='button'
              >
                Create
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
            aria-label={toggleLabel}
            className={styles.toggle}
            data-testid='recurring-item-toggle'
            onClick={() => setExpanded(!expanded)}
            type='button'
          >
            {expanded && <ChevronUp />}
            {!expanded && <ChevronDown />}
          </button>
        </div>
        {expanded && <RecurringItemCharges detailsPath={detailsPath} name={name} recurring={recurring} />}
      </div>
    </li>
  );
}
