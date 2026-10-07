import { Fragment, useEffect, useMemo, useState } from 'react';
import { differenceInCalendarDays, differenceInCalendarMonths, isBefore, isThisYear, startOfToday } from 'date-fns';
import {
  CalendarSync,
  ChevronDown,
  ChevronUp,
  HeartCrack,
  Info,
  Layers,
  Receipt,
  Repeat,
  TriangleAlert,
} from 'lucide-react';
import { rrulestr } from 'rrule';
import { Link } from 'wouter';

import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import Typography from '@monetr/interface/components/Typography';
import { useFundingSchedules } from '@monetr/interface/hooks/useFundingSchedules';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useRecurringTransactionHistory } from '@monetr/interface/hooks/useRecurringTransactionHistory';
import { useRecurringTransactions } from '@monetr/interface/hooks/useRecurringTransactions';
import { useSpending } from '@monetr/interface/hooks/useSpending';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { showNewExpenseModal } from '@monetr/interface/modals/NewExpenseModal';
import { showNewFundingModal } from '@monetr/interface/modals/NewFundingModal';
import type Spending from '@monetr/interface/models/Spending';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import { TransactionRecurringWindow } from '@monetr/interface/models/TransactionRecurring';
import { AmountType } from '@monetr/interface/util/amounts';
import capitalize from '@monetr/interface/util/capitalize';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './recurring.module.scss';

type RecurringTab = 'charges' | 'deposits' | 'ended';

export default function Recurring(): React.JSX.Element {
  const {
    data: recurring,
    isLoading,
    isError,
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage,
  } = useRecurringTransactions();
  const [tab, setTab] = useState<RecurringTab>('charges');

  // the page groups and totals everything up, so it needs every page and not just the first one
  useEffect(() => {
    if (hasNextPage && !isFetchingNextPage) {
      void fetchNextPage();
    }
  }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

  const tabs = useMemo(() => {
    const items = recurring ?? [];
    return {
      charges: items.filter(item => !item.ended && item.direction === 'debit'),
      deposits: items.filter(item => !item.ended && item.direction === 'credit'),
      ended: items.filter(item => item.ended),
    };
  }, [recurring]);

  if (isLoading) {
    return (
      <div className={styles.centerState}>
        <Typography size='5xl'>One moment...</Typography>
      </div>
    );
  }

  if (isError) {
    return (
      <div className={styles.centerState}>
        <HeartCrack className={styles.errorIcon} />
        <Typography size='5xl'>Something isn&apos;t right...</Typography>
        <Typography size='2xl'>We weren&apos;t able to retrieve recurring transactions at this time...</Typography>
      </div>
    );
  }

  return (
    <Fragment>
      <MTopNavigation icon={Repeat} title='Recurring' />
      <div className={styles.body}>
        {recurring?.length === 0 ? (
          <div className={styles.empty}>
            <Repeat className={styles.emptyIcon} />
            <Typography align='center' color='subtle' size='xl'>
              monetr hasn&apos;t found anything recurring yet...
            </Typography>
            <Typography align='center' color='subtle' size='lg'>
              {FOOTNOTE}
            </Typography>
          </div>
        ) : (
          <Fragment>
            <div className={styles.tabs} role='tablist'>
              <TabButton count={tabs.charges.length} onClick={() => setTab('charges')} selected={tab === 'charges'}>
                Charges
              </TabButton>
              <TabButton count={tabs.deposits.length} onClick={() => setTab('deposits')} selected={tab === 'deposits'}>
                Deposits
              </TabButton>
              <TabButton count={tabs.ended.length} onClick={() => setTab('ended')} selected={tab === 'ended'}>
                Ended
              </TabButton>
            </div>
            {tab === 'ended' ? <EndedList items={tabs.ended} /> : <UpcomingList items={tabs[tab]} />}
            <div className={styles.footnote}>
              <Info />
              <Typography color='subtle' size='sm'>
                {FOOTNOTE}
              </Typography>
            </div>
          </Fragment>
        )}
      </div>
    </Fragment>
  );
}

const FOOTNOTE =
  'monetr looks for charges that repeat on a schedule once it has seen at least three of them. New ones can take a day or two to show up here.';

interface TabButtonProps {
  children: React.ReactNode;
  count: number;
  selected: boolean;
  onClick: () => void;
}

function TabButton({ children, count, selected, onClick }: TabButtonProps): React.JSX.Element {
  return (
    <button aria-selected={selected} className={styles.tab} onClick={onClick} role='tab' type='button'>
      {children}
      <span className={styles.tabCount}>{count}</span>
    </button>
  );
}

// useNextFunding gives back the earliest date any funding schedule on this account will contribute next
function useNextFunding(): Date | null {
  const { data: funding } = useFundingSchedules();
  return useMemo(
    () =>
      (funding ?? []).reduce<Date | null>(
        (earliest, item) => (!earliest || isBefore(item.nextRecurrence, earliest) ? item.nextRecurrence : earliest),
        null,
      ),
    [funding],
  );
}

function useFormatters() {
  const { timezone, inTimezone } = useTimezone();
  const { data: locale } = useLocale();
  const { data: localeCurrency } = useLocaleCurrency();
  if (!locale || !localeCurrency) {
    return null;
  }

  const today = startOfToday({ in: inTimezone });
  return {
    today,
    amount: (amount: number) => localeCurrency.formatAmount(Math.abs(amount), AmountType.Stored),
    // only include the year when it isn't this year, otherwise something next february reads like this february
    date: (date: Date, options: Intl.DateTimeFormatOptions = {}) =>
      new Intl.DateTimeFormat(locale.code, {
        month: 'short',
        day: 'numeric',
        year: isThisYear(date, { in: inTimezone }) ? undefined : 'numeric',
        timeZone: timezone,
        ...options,
      }).format(date),
    daysFromToday: (date: Date) => differenceInCalendarDays(inTimezone(date), today),
    relative: (date: Date) => {
      const format = new Intl.RelativeTimeFormat(locale.code, { numeric: 'auto' });
      const days = differenceInCalendarDays(inTimezone(date), today);
      // days get hard to picture past a month or two, so switch over to months for things that are far out
      if (Math.abs(days) >= 60) {
        return format.format(differenceInCalendarMonths(inTimezone(date), today), 'month');
      }
      return format.format(days, 'day');
    },
  };
}

// getShortfall is how much of the next charge the linked expense won't cover. Only when the expense is actually behind,
// one that will catch up by the time the charge comes in isn't short. Nothing until the expense has loaded either, so
// it doesn't flash a shortfall for the whole amount in the meantime.
function getShortfall(item: TransactionRecurring, spending: Spending | undefined): number {
  if (item.ended || !spending?.isBehind) {
    return 0;
  }
  return Math.max(0, Math.abs(item.lastAmount) - spending.currentAmount);
}

interface ListProps {
  items: Array<TransactionRecurring>;
}

function UpcomingList({ items }: ListProps): React.JSX.Element | null {
  const nextFunding = useNextFunding();
  const format = useFormatters();
  if (!format) {
    return null;
  }

  if (items.length === 0) {
    return (
      <Typography className={styles.emptyTab} color='subtle'>
        Nothing here yet.
      </Typography>
    );
  }

  const late = items.filter(item => isBefore(item.next, format.today));
  const beforeFunding = nextFunding
    ? items.filter(item => !isBefore(item.next, format.today) && isBefore(item.next, nextFunding))
    : [];
  const later = items.filter(item => !late.includes(item) && !beforeFunding.includes(item));
  const sum = (group: Array<TransactionRecurring>) =>
    format.amount(group.reduce((total, item) => total + Math.abs(item.lastAmount), 0));

  return (
    <div className={styles.list}>
      {late.length > 0 && (
        <Fragment>
          <div className={styles.groupHeader} data-late>
            <span className={styles.groupTitle}>
              <TriangleAlert />
              Late
            </span>
            <span>{sum(late)}</span>
          </div>
          {late.map(item => (
            <RecurringRow item={item} key={item.transactionRecurringId.toString()} />
          ))}
        </Fragment>
      )}
      {beforeFunding.length > 0 && nextFunding && (
        <Fragment>
          <div className={styles.groupHeader}>
            <span>Before {format.date(nextFunding)} funding</span>
            <span>{sum(beforeFunding)}</span>
          </div>
          {beforeFunding.map(item => (
            <RecurringRow item={item} key={item.transactionRecurringId.toString()} />
          ))}
        </Fragment>
      )}
      {later.length > 0 && (
        <Fragment>
          <div className={styles.groupHeader}>
            {/* without a funding schedule theres nothing to split on, so everything is just upcoming */}
            <span>{late.length > 0 || beforeFunding.length > 0 ? 'Later' : 'Upcoming'}</span>
          </div>
          {later.map(item => (
            <RecurringRow item={item} key={item.transactionRecurringId.toString()} />
          ))}
        </Fragment>
      )}
    </div>
  );
}

function EndedList({ items }: ListProps): React.JSX.Element {
  if (items.length === 0) {
    return (
      <Typography className={styles.emptyTab} color='subtle'>
        Nothing has stopped repeating.
      </Typography>
    );
  }

  return (
    <div className={styles.list}>
      {items.map(item => (
        <RecurringRow item={item} key={item.transactionRecurringId.toString()} />
      ))}
    </div>
  );
}

interface RecurringRowProps {
  item: TransactionRecurring;
}

function RecurringRow({ item }: RecurringRowProps): React.JSX.Element | null {
  const format = useFormatters();
  const [expanded, setExpanded] = useState(false);
  if (!format) {
    return null;
  }

  const name = item.transactionCluster?.name ?? 'Recurring';
  const memo = item.transactionCluster?.originalMemo;
  const detailsUrl = `/bank/${item.bankAccountId}/recurring/${item.transactionRecurringId}/details`;
  const daysLate = -format.daysFromToday(item.next);
  const cadence = getCadence(item);

  return (
    <div className={styles.rowGroup} data-expanded={expanded}>
      <div className={styles.row} data-testid={item.transactionRecurringId.toString()}>
        <Link aria-label={name} className={styles.rowLink} to={detailsUrl} />
        <MerchantIcon className={styles.rowIcon} name={name} />
        <div className={styles.rowName}>
          <span className={styles.rowTitle}>
            <span className={styles.rowTitleName}>{name}</span>
            {/* desktop has the cadence in the date column, mobile tucks it in after the name */}
            <span className={styles.rowTitleCadence}>· {cadence}</span>
          </span>
          {/* the raw text from the bank, so its easy to tell which charges on a statement this is */}
          {memo && (
            <span className={styles.rowMemo} title={memo}>
              {memo}
            </span>
          )}
          <RecurringStatus item={item} />
        </div>
        <div className={styles.rowDate}>
          {item.ended ? (
            <Fragment>
              <span className={styles.rowDateValue}>{format.date(item.last)}</span>
              <span className={styles.rowSubtle}>{cadence} · last seen</span>
            </Fragment>
          ) : daysLate > 0 ? (
            <Fragment>
              <span className={styles.lateBadge}>{daysLate === 1 ? '1 day late' : `${daysLate} days late`}</span>
              <span className={styles.rowSubtle}>
                {cadence} · was due {format.date(item.next)}
              </span>
            </Fragment>
          ) : (
            <Fragment>
              <span className={styles.rowDateValue}>{format.date(item.next)}</span>
              <span className={styles.rowSubtle}>
                {cadence} · {format.relative(item.next)}
              </span>
            </Fragment>
          )}
        </div>
        <div className={styles.rowBudget}>
          <RecurringBudget item={item} />
        </div>
        <RecurringAmount item={item} />
        <button
          aria-expanded={expanded}
          aria-label={`${expanded ? 'Hide' : 'Show'} recent ${name} charges`}
          className={styles.rowToggle}
          onClick={() => setExpanded(!expanded)}
          type='button'
        >
          {expanded ? <ChevronUp /> : <ChevronDown />}
        </button>
      </div>
      {expanded && <RecentCharges detailsUrl={detailsUrl} item={item} name={name} />}
    </div>
  );
}

interface RecentChargesProps {
  item: TransactionRecurring;
  name: string;
  detailsUrl: string;
}

// RecentCharges is the peek you get when expanding a row, the last few charges with what they actually looked like on
// the statement. Only fetched once the row is opened.
function RecentCharges({ item, name, detailsUrl }: RecentChargesProps): React.JSX.Element | null {
  const { transactions, seen, isLoading } = useRecurringTransactionHistory(item);
  const format = useFormatters();
  const { inTimezone } = useTimezone();
  const { data: locale } = useLocale();
  if (!format || !locale) {
    return null;
  }

  const since = new Intl.DateTimeFormat(locale.code, { month: 'short', year: 'numeric' }).format(
    inTimezone(item.first),
  );

  return (
    <div className={styles.recent}>
      <div className={styles.recentHeader}>
        <Layers />
        <span>Similar transactions group</span>
        <strong>{name}</strong>
        <span>
          · {seen === 1 ? '1 charge' : `${seen} charges`} since {since}
        </span>
        <Link className={styles.recentLink} to={detailsUrl}>
          View details
        </Link>
      </div>
      {isLoading ? (
        <span className={styles.rowSubtle}>Loading...</span>
      ) : (
        <div className={styles.recentList}>
          {transactions.slice(0, 3).map(transaction => (
            <Fragment key={transaction.transactionId}>
              <span className={styles.recentDate}>{format.date(transaction.date)}</span>
              <span className={styles.recentMemo} title={transaction.originalName}>
                {transaction.originalName}
              </span>
              <span className={styles.recentAmount} data-addition={transaction.getIsAddition()}>
                {format.amount(transaction.amount)}
              </span>
            </Fragment>
          ))}
        </div>
      )}
    </div>
  );
}

// getCadence is the short version of the schedule for the date column, the full ruleset text is on the details page
function getCadence(item: TransactionRecurring): string {
  switch (item.window) {
    case TransactionRecurringWindow.FirstAndFifteenth:
    case TransactionRecurringWindow.FifteenthAndLast:
      return 'Twice a month';
    case TransactionRecurringWindow.Weekly:
      return 'Weekly';
    case TransactionRecurringWindow.BiWeekly:
      return 'Every 2 weeks';
    case TransactionRecurringWindow.Monthly:
      return 'Monthly';
    case TransactionRecurringWindow.BiMonthly:
      return 'Every 2 months';
    case TransactionRecurringWindow.Quarterly:
      return 'Quarterly';
    case TransactionRecurringWindow.SemiYearly:
      return 'Every 6 months';
    case TransactionRecurringWindow.Yearly:
      return 'Yearly';
    default:
      return capitalize(rrulestr(item.ruleset).toText());
  }
}

function RecurringBudget({ item }: RecurringRowProps): React.JSX.Element | null {
  const format = useFormatters();
  // the list doesn't embed the expense, it gets looked up so it comes from the same cache as the rest of the app
  const { data: spending } = useSpending(item.spendingId);
  if (!format) {
    return null;
  }

  if (item.direction === 'debit') {
    if (item.spendingId) {
      const shortfall = getShortfall(item, spending);
      return (
        <Fragment>
          <span className={styles.budgetIcon}>
            <Receipt />
          </span>
          <span className={styles.budgetName}>{spending?.name}</span>
          {shortfall > 0 && <span className={styles.budgetShort}>short {format.amount(shortfall)}</span>}
        </Fragment>
      );
    }

    // the button already says theres no expense yet, so only spell it out when theres nothing to create anymore
    if (item.ended) {
      return <span className={styles.rowSubtle}>Not budgeted</span>;
    }

    return (
      <button className={styles.budgetButton} onClick={() => showNewExpenseModal({ recurring: item })} type='button'>
        Create expense
      </button>
    );
  }

  if (item.fundingSchedule) {
    return (
      <Fragment>
        <span className={styles.budgetIcon}>
          <CalendarSync />
        </span>
        <span className={styles.budgetName}>{item.fundingSchedule.name}</span>
      </Fragment>
    );
  }

  if (item.ended) {
    return <span className={styles.rowSubtle}>Not funding anything</span>;
  }

  return (
    <button className={styles.budgetButton} onClick={() => showNewFundingModal({ recurring: item })} type='button'>
      Create funding schedule
    </button>
  );
}

function RecurringAmount({ item }: RecurringRowProps): React.JSX.Element | null {
  const format = useFormatters();
  if (!format) {
    return null;
  }

  // a price change shows up as two amounts, so only call it variable once theres more than that going on
  const amounts = Object.keys(item.amounts).map(amount => Math.abs(Number(amount)));
  const variable = amounts.length >= 3;

  return (
    <div className={styles.rowAmount}>
      <span>
        {variable && '~'}
        {format.amount(item.lastAmount)}
      </span>
      {variable && (
        <span className={mergeClasses(styles.rowSubtle, styles.rowAmountRange)}>
          {format.amount(Math.min(...amounts))} to {format.amount(Math.max(...amounts))}
        </span>
      )}
      {/* on mobile theres no date column, so the date tucks in under the amount instead */}
      <span className={styles.rowAmountDate}>{format.date(item.ended ? item.last : item.next)}</span>
    </div>
  );
}

// RecurringStatus is the one line under the name on mobile, it stands in for the budget column which doesnt fit there.
function RecurringStatus({ item }: RecurringRowProps): React.JSX.Element | null {
  const format = useFormatters();
  const { data: spending } = useSpending(item.spendingId);
  if (!format) {
    return null;
  }

  if (item.direction === 'debit') {
    if (!item.spendingId) {
      return (
        <span className={styles.rowStatus} data-warning={!item.ended}>
          Not budgeted
        </span>
      );
    }

    const shortfall = getShortfall(item, spending);
    return (
      <span className={styles.rowStatus}>
        <Receipt />
        <span className={styles.rowStatusName}>{spending?.name}</span>
        {shortfall > 0 && <span className={styles.budgetShort}>· short {format.amount(shortfall)}</span>}
      </span>
    );
  }

  if (!item.fundingSchedule) {
    return <span className={styles.rowStatus}>Not funding anything</span>;
  }

  return (
    <span className={styles.rowStatus}>
      <CalendarSync />
      <span className={styles.rowStatusName}>{item.fundingSchedule.name}</span>
    </span>
  );
}
