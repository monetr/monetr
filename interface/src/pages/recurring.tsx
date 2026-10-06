import { Fragment, useEffect, useMemo, useState } from 'react';
import { differenceInCalendarDays, differenceInCalendarMonths, isBefore, isThisYear, startOfToday } from 'date-fns';
import { CalendarSync, ChevronRight, HeartCrack, Receipt, Repeat, TriangleAlert } from 'lucide-react';
import { rrulestr } from 'rrule';
import { Link } from 'wouter';

import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import Typography from '@monetr/interface/components/Typography';
import { useFundingSchedules } from '@monetr/interface/hooks/useFundingSchedules';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useRecurringTransactions } from '@monetr/interface/hooks/useRecurringTransactions';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { showNewExpenseModal } from '@monetr/interface/modals/NewExpenseModal';
import { showNewFundingModal } from '@monetr/interface/modals/NewFundingModal';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
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
            <NextFundingSummary charges={tabs.charges} />
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
            <Typography className={styles.footnote} color='subtle' size='sm'>
              {FOOTNOTE}
            </Typography>
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

// getShortfall is how much of the next charge the linked expense won't cover, or all of it when there isn't one
function getShortfall(item: TransactionRecurring): number {
  const amount = Math.abs(item.lastAmount);
  if (!item.spending) {
    return amount;
  }
  return Math.max(0, amount - item.spending.currentAmount);
}

interface NextFundingSummaryProps {
  charges: Array<TransactionRecurring>;
}

function NextFundingSummary({ charges }: NextFundingSummaryProps): React.JSX.Element | null {
  const nextFunding = useNextFunding();
  const format = useFormatters();
  if (!nextFunding || !format) {
    return null;
  }

  const upcoming = charges.filter(item => !isBefore(item.next, format.today) && isBefore(item.next, nextFunding));
  const total = upcoming.reduce((sum, item) => sum + Math.abs(item.lastAmount), 0);
  const uncovered = upcoming.reduce((sum, item) => sum + getShortfall(item), 0);

  return (
    <section className={styles.summary}>
      <span className={styles.summaryEyebrow}>
        Before your next funding on {format.date(nextFunding, { weekday: 'long' })}
      </span>
      {upcoming.length > 0 && (
        // mobile gets the total up front and the rest underneath, theres not enough room for the whole sentence
        <div className={styles.summaryMobile}>
          <span className={styles.summaryTotal}>
            {format.amount(total)}{' '}
            <span>expected · {upcoming.length === 1 ? '1 charge' : `${upcoming.length} charges`}</span>
          </span>
          <span className={styles.summaryDetail}>
            {uncovered > 0 ? (
              <Fragment>
                <strong>{format.amount(uncovered)}</strong> isn&apos;t budgeted and will come out of Free-To-Use
              </Fragment>
            ) : (
              'All of it is covered by your expenses'
            )}
          </span>
        </div>
      )}
      {upcoming.length === 0 ? (
        <p className={styles.summaryText}>Nothing recurring is expected before then.</p>
      ) : (
        <p className={mergeClasses(styles.summaryText, styles.summaryDesktop)}>
          <strong>{format.amount(total)}</strong> in {upcoming.length === 1 ? '1 charge' : `${upcoming.length} charges`}{' '}
          is expected.{' '}
          {uncovered > 0 ? (
            <Fragment>
              <strong data-warning>{format.amount(uncovered)}</strong> of that isn&apos;t covered by an expense and will
              come out of Free-To-Use.
            </Fragment>
          ) : (
            'All of it is covered by your expenses.'
          )}
        </p>
      )}
    </section>
  );
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
  if (!format) {
    return null;
  }

  const name = item.transactionCluster?.name ?? 'Recurring';
  const detailsUrl = `/bank/${item.bankAccountId}/recurring/${item.transactionRecurringId}/details`;
  const daysLate = -format.daysFromToday(item.next);

  return (
    <div className={styles.row} data-testid={item.transactionRecurringId.toString()}>
      <Link aria-label={name} className={styles.rowLink} to={detailsUrl} />
      <MerchantIcon className={styles.rowIcon} name={name} />
      <div className={styles.rowName}>
        <span className={styles.rowTitle}>{name}</span>
        <span className={styles.rowSubtitle}>{capitalize(rrulestr(item.ruleset).toText())}</span>
        {item.transactionCluster?.originalMemo && (
          <span className={styles.rowMemo} title={item.transactionCluster.originalMemo}>
            {item.transactionCluster.originalMemo}
          </span>
        )}
        <RecurringStatus item={item} />
      </div>
      <div className={styles.rowDate}>
        {item.ended ? (
          <Fragment>
            <span className={styles.rowDateValue}>{format.date(item.last)}</span>
            <span className={styles.rowSubtle}>Last seen</span>
          </Fragment>
        ) : daysLate > 0 ? (
          <Fragment>
            <span className={styles.lateBadge}>{daysLate === 1 ? '1 day late' : `${daysLate} days late`}</span>
            <span className={styles.rowSubtle}>Expected {format.date(item.next)}</span>
          </Fragment>
        ) : (
          <Fragment>
            <span className={styles.rowDateValue}>{format.date(item.next)}</span>
            <span className={styles.rowSubtle}>{format.relative(item.next)}</span>
          </Fragment>
        )}
      </div>
      <div className={styles.rowBudget}>
        <RecurringBudget item={item} />
      </div>
      <RecurringAmount item={item} />
      <ChevronRight className={styles.rowChevron} />
    </div>
  );
}

function RecurringBudget({ item }: RecurringRowProps): React.JSX.Element | null {
  const format = useFormatters();
  if (!format) {
    return null;
  }

  if (item.direction === 'debit') {
    if (item.spending) {
      const shortfall = item.ended ? 0 : getShortfall(item);
      return (
        <Fragment>
          <span className={styles.budgetIcon}>
            <Receipt />
          </span>
          <span className={styles.budgetName}>{item.spending.name}</span>
          {shortfall > 0 && <span className={styles.budgetShort}>short {format.amount(shortfall)}</span>}
        </Fragment>
      );
    }

    return (
      <Fragment>
        <span className={styles.rowSubtle}>Not budgeted</span>
        {!item.ended && (
          <button
            className={styles.budgetButton}
            onClick={() => showNewExpenseModal({ recurring: item })}
            type='button'
          >
            Create expense
          </button>
        )}
      </Fragment>
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

  return (
    <Fragment>
      <span className={styles.rowSubtle}>Not funding anything</span>
      {!item.ended && (
        <button className={styles.budgetButton} onClick={() => showNewFundingModal({ recurring: item })} type='button'>
          Create funding schedule
        </button>
      )}
    </Fragment>
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
  if (!format) {
    return null;
  }

  if (item.direction === 'debit') {
    if (!item.spending) {
      return (
        <span className={styles.rowStatus} data-warning={!item.ended}>
          Not budgeted
        </span>
      );
    }

    const shortfall = item.ended ? 0 : getShortfall(item);
    return (
      <span className={styles.rowStatus}>
        <Receipt />
        <span className={styles.rowStatusName}>{item.spending.name}</span>
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
