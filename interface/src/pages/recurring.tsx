import { Fragment, useEffect, useState } from 'react';
import { format, isBefore, isThisYear, startOfToday } from 'date-fns';
import { HeartCrack, Info, Repeat, TriangleAlert } from 'lucide-react';

import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import RecurringItem from '@monetr/interface/components/recurring/RecurringItem';
import Typography from '@monetr/interface/components/Typography';
import { useFundingSchedules } from '@monetr/interface/hooks/useFundingSchedules';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useRecurringTransactions } from '@monetr/interface/hooks/useRecurringTransactions';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import { AmountType } from '@monetr/interface/util/amounts';

import styles from './recurring.module.scss';

type Tab = 'charges' | 'deposits' | 'ended';

export default function Recurring(): React.JSX.Element {
  const {
    data: recurring,
    isLoading,
    isError,
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage,
  } = useRecurringTransactions();
  const { data: fundingSchedules } = useFundingSchedules();
  const { data: locale } = useLocaleCurrency();
  const { inTimezone } = useTimezone();
  const [tab, setTab] = useState<Tab>('charges');

  // The page groups and totals everything, so keep loading until we have every page instead of just the first one.
  useEffect(() => {
    if (hasNextPage && !isFetchingNextPage) {
      fetchNextPage();
    }
  }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

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

  if (!recurring || !locale) {
    return <Fragment />;
  }

  const charges = recurring.filter(item => !item.ended && item.direction === 'debit');
  const deposits = recurring.filter(item => !item.ended && item.direction === 'credit');
  const ended = recurring.filter(item => item.ended);

  let items = charges;
  if (tab === 'deposits') {
    items = deposits;
  } else if (tab === 'ended') {
    items = ended;
  }

  // Split upcoming ones by whether they land before the next funding. Ended ones aren't upcoming so they don't get
  // split up at all.
  const today = startOfToday({ in: inTimezone });
  const nextFunding = fundingSchedules
    ?.map(item => item.nextRecurrence)
    .sort((a, b) => a.getTime() - b.getTime())
    .at(0);
  const late = tab === 'ended' ? [] : items.filter(item => isBefore(item.next, today));
  const beforeFunding =
    tab === 'ended' || !nextFunding
      ? []
      : items.filter(item => !isBefore(item.next, today) && isBefore(item.next, nextFunding));
  const later = items.filter(item => !late.includes(item) && !beforeFunding.includes(item));

  function sum(group: Array<TransactionRecurring>): string {
    const total = group.reduce((total, item) => total + Math.abs(item.lastAmount), 0);
    return locale?.formatAmount(total, AmountType.Stored) ?? '';
  }

  // If nothing got split off then there isn't really anything to call these "later" than.
  let laterTitle = 'Later';
  if (late.length === 0 && beforeFunding.length === 0) {
    laterTitle = 'Upcoming';
  }

  return (
    <Fragment>
      <MTopNavigation icon={Repeat} title='Recurring' />
      <div className={styles.content}>
        {recurring.length === 0 && <EmptyState />}
        {recurring.length > 0 && (
          <Fragment>
            <div className={styles.tabs} role='tablist'>
              <button
                aria-selected={tab === 'charges'}
                className={styles.tab}
                onClick={() => setTab('charges')}
                role='tab'
                type='button'
              >
                Charges<span className={styles.tabCount}>{charges.length}</span>
              </button>
              <button
                aria-selected={tab === 'deposits'}
                className={styles.tab}
                onClick={() => setTab('deposits')}
                role='tab'
                type='button'
              >
                Deposits<span className={styles.tabCount}>{deposits.length}</span>
              </button>
              <button
                aria-selected={tab === 'ended'}
                className={styles.tab}
                onClick={() => setTab('ended')}
                role='tab'
                type='button'
              >
                Ended<span className={styles.tabCount}>{ended.length}</span>
              </button>
            </div>

            {items.length === 0 && (
              <Typography className={styles.emptyTab} color='subtle'>
                {tab === 'ended' ? 'Nothing has stopped repeating.' : 'Nothing here yet.'}
              </Typography>
            )}

            <ul className={styles.list}>
              {late.length > 0 && (
                <li className={styles.groupHeader} data-late='true'>
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

              {beforeFunding.length > 0 && nextFunding && (
                <li className={styles.groupHeader}>
                  <span>
                    Before{' '}
                    {isThisYear(nextFunding)
                      ? format(inTimezone(nextFunding), 'MMM d')
                      : format(inTimezone(nextFunding), 'MMM d, yyyy')}{' '}
                    funding
                  </span>
                  <span>{sum(beforeFunding)}</span>
                </li>
              )}
              {beforeFunding.map(item => (
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

            <div className={styles.footnote}>
              <Info />
              <Typography color='subtle' size='sm'>
                monetr looks for charges that repeat on a schedule once it has seen at least three of them. New ones can
                take a day or two to show up here.
              </Typography>
            </div>
          </Fragment>
        )}
      </div>
    </Fragment>
  );
}

function EmptyState(): React.JSX.Element {
  return (
    <div className={styles.empty}>
      <Repeat className={styles.emptyIcon} />
      <Typography align='center' color='subtle' size='xl'>
        monetr hasn&apos;t found anything recurring yet...
      </Typography>
      <Typography align='center' color='subtle' size='lg'>
        monetr looks for charges that repeat on a schedule once it has seen at least three of them. New ones can take a
        day or two to show up here.
      </Typography>
    </div>
  );
}
