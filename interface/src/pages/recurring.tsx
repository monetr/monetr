import { Fragment } from 'react';
import { isBefore, startOfToday } from 'date-fns';
import { HeartCrack, Info, Repeat, TriangleAlert } from 'lucide-react';
import { useSearchParams } from 'wouter';

import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import RecurringItem from '@monetr/interface/components/recurring/RecurringItem';
import Typography from '@monetr/interface/components/Typography';
import { useInfiniteScroll } from '@monetr/interface/hooks/useInfiniteScroll';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useRecurringTransactions } from '@monetr/interface/hooks/useRecurringTransactions';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import { AmountType } from '@monetr/interface/util/amounts';

import styles from './recurring.module.scss';

type Tab = 'charges' | 'deposits' | 'ended';

export default function Recurring(): React.JSX.Element | null {
  const [searchParams, setSearchParams] = useSearchParams();

  // The tab lives in the url so coming back to this page, like from a details page, keeps you on the same tab
  let tab: Tab = 'charges';
  const tabParam = searchParams.get('tab');
  if (tabParam === 'deposits' || tabParam === 'ended') {
    tab = tabParam;
  }

  // Only ask the server for whatever the tab is showing
  let direction: TransactionRecurring['direction'] | undefined = 'debit';
  let ended = false;
  if (tab === 'deposits') {
    direction = 'credit';
  } else if (tab === 'ended') {
    direction = undefined;
    ended = true;
  }

  const {
    data: recurring,
    isLoading,
    isFetching,
    isError,
    hasNextPage,
    fetchNextPage,
  } = useRecurringTransactions({
    direction: direction,
    ended: ended,
  });
  const { data: locale } = useLocaleCurrency();
  const { inTimezone } = useTimezone();
  const [sentryRef] = useInfiniteScroll({
    loading: isFetching,
    hasNextPage,
    onLoadMore: fetchNextPage,
    disabled: isError,
    // Start loading the next page a bit before the bottom actually shows up
    rootMargin: '0px 0px 700px 0px',
  });

  if (isError) {
    return (
      <div className={styles.centerState}>
        <HeartCrack className={styles.errorIcon} />
        <Typography size='5xl'>Something isn&apos;t right...</Typography>
        <Typography size='2xl'>We weren&apos;t able to retrieve recurring transactions at this time...</Typography>
      </div>
    );
  }

  if (!locale) {
    return null;
  }

  const items = recurring ?? [];

  // Anything that was expected before today is late. The server sorts by when they're expected next so these are
  // always on the first page. Ended ones aren't upcoming so they don't get split up at all.
  const today = startOfToday({
    in: inTimezone,
  });
  let late: Array<TransactionRecurring> = [];
  if (tab !== 'ended') {
    late = items.filter(item => isBefore(item.next, today));
  }
  const later = items.filter(item => !late.includes(item));

  // If nothing is late then there isn't really anything to call these later than
  let laterTitle = 'Later';
  if (late.length === 0) {
    laterTitle = 'Upcoming';
  }

  let message = 'Nothing here yet...';
  if (isLoading) {
    message = 'Loading...';
  } else if (tab === 'ended') {
    message = 'Nothing has stopped repeating...';
  }

  function changeTab(next: Tab) {
    setSearchParams({
      tab: next,
    });
  }

  function sum(group: Array<TransactionRecurring>): string {
    const total = group.reduce((total, item) => total + Math.abs(item.lastAmount), 0);
    return locale?.formatAmount(total, AmountType.Stored) ?? '';
  }

  return (
    <Fragment>
      <MTopNavigation icon={Repeat} title='Recurring' />
      <div className={styles.content}>
        <div className={styles.tabs} role='tablist'>
          <button
            aria-selected={tab === 'charges'}
            className={styles.tab}
            data-testid='recurring-tab-charges'
            onClick={() => changeTab('charges')}
            role='tab'
            type='button'
          >
            Charges
          </button>
          <button
            aria-selected={tab === 'deposits'}
            className={styles.tab}
            data-testid='recurring-tab-deposits'
            onClick={() => changeTab('deposits')}
            role='tab'
            type='button'
          >
            Deposits
          </button>
          <button
            aria-selected={tab === 'ended'}
            className={styles.tab}
            data-testid='recurring-tab-ended'
            onClick={() => changeTab('ended')}
            role='tab'
            type='button'
          >
            Ended
          </button>
        </div>

        {items.length === 0 && (
          <Typography className={styles.message} color='subtle' data-testid='recurring-empty'>
            {message}
          </Typography>
        )}

        <ul className={styles.list}>
          {late.length > 0 && (
            <li className={styles.groupHeader} data-late='true' data-testid='recurring-group-late'>
              <span className={styles.groupTitle}>
                <TriangleAlert />
                Late
              </span>
              <span>{sum(late)}</span>
            </li>
          )}
          {late.map(item => (
            <RecurringItem key={item.transactionRecurringId} recurring={item} />
          ))}

          {later.length > 0 && tab !== 'ended' && (
            <li className={styles.groupHeader}>
              <span>{laterTitle}</span>
            </li>
          )}
          {later.map(item => (
            <RecurringItem key={item.transactionRecurringId} recurring={item} />
          ))}
        </ul>

        {hasNextPage && (
          <div ref={sentryRef}>
            <Typography className={styles.message} color='subtle'>
              Loading...
            </Typography>
          </div>
        )}

        <div className={styles.footnote}>
          <Info />
          <Typography color='subtle' size='sm'>
            monetr looks for charges that repeat on a schedule once it has seen at least three of them. New ones can
            take a day or two to show up here.
          </Typography>
        </div>
      </div>
    </Fragment>
  );
}
