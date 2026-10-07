import { useState } from 'react';

import RecurringChargeItem from '@monetr/interface/components/recurring/RecurringChargeItem';
import Typography from '@monetr/interface/components/Typography';
import { useInfiniteScroll } from '@monetr/interface/hooks/useInfiniteScroll';
import { useSimilarTransactions } from '@monetr/interface/hooks/useSimilarTransactions';
import { useTransactionsForRecurring } from '@monetr/interface/hooks/useTransactionsForRecurring';
import type { ID } from '@monetr/interface/models/ID';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';

import styles from './RecurringChargeList.module.scss';

type Tab = 'schedule' | 'all';

export interface RecurringChargeListProps {
  recurring: TransactionRecurring;
}

export default function RecurringChargeList(props: RecurringChargeListProps): React.JSX.Element {
  const [tab, setTab] = useState<Tab>('schedule');

  // Only load whichever tab is showing, plenty of people will never look at the all tab. On schedule is everything
  // monetr matched to this recurring transaction, all is the whole similar transactions group.
  let transactionRecurringId: ID<TransactionRecurring> | undefined;
  let transactionClusterId: string | undefined;
  if (tab === 'schedule') {
    transactionRecurringId = props.recurring.transactionRecurringId;
  } else {
    transactionClusterId = props.recurring.transactionClusterId;
  }
  const onSchedule = useTransactionsForRecurring(transactionRecurringId, 10);
  const similar = useSimilarTransactions(transactionClusterId);

  let query = onSchedule;
  if (tab === 'all') {
    query = similar;
  }
  const transactions = query.data ?? [];
  const [sentryRef] = useInfiniteScroll({
    loading: query.isFetching,
    hasNextPage: query.hasNextPage,
    onLoadMore: query.fetchNextPage,
    disabled: query.isError,
    // Start loading the next page a bit before the bottom actually shows up
    rootMargin: '0px 0px 700px 0px',
  });

  let heading = 'Deposits';
  if (props.recurring.direction === 'debit') {
    heading = 'Charges';
  }

  return (
    <section className={styles.root}>
      <div className={styles.header}>
        <Typography color='emphasis' component='h3' size='md' weight='semibold'>
          {heading}
        </Typography>
        <div className={styles.tabs} role='tablist'>
          <button
            aria-selected={tab === 'schedule'}
            className={styles.tab}
            onClick={() => setTab('schedule')}
            role='tab'
            type='button'
          >
            On Schedule
          </button>
          <button
            aria-selected={tab === 'all'}
            className={styles.tab}
            onClick={() => setTab('all')}
            role='tab'
            type='button'
          >
            All
          </button>
        </div>
      </div>

      <ul className={styles.list}>
        {transactions.map(item => (
          <RecurringChargeItem key={item.transactionId} recurring={props.recurring} transaction={item} />
        ))}
      </ul>

      {!query.isLoading && transactions.length === 0 && (
        <Typography className={styles.message} color='subtle' size='sm'>
          Nothing here...
        </Typography>
      )}
      {query.hasNextPage && (
        <div ref={sentryRef}>
          <Typography className={styles.message} color='subtle' size='sm'>
            Loading...
          </Typography>
        </div>
      )}
    </section>
  );
}
