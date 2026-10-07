import { useState } from 'react';

import RecurringChargeItem from '@monetr/interface/components/recurring/RecurringChargeItem';
import Typography from '@monetr/interface/components/Typography';
import { useRecurringTransactionHistory } from '@monetr/interface/hooks/useRecurringTransactionHistory';
import { useSimilarTransactions } from '@monetr/interface/hooks/useSimilarTransactions';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';

import styles from './RecurringChargeList.module.scss';

type Tab = 'schedule' | 'all';

export interface RecurringChargeListProps {
  recurring: TransactionRecurring;
}

export default function RecurringChargeList(props: RecurringChargeListProps): React.JSX.Element {
  const [tab, setTab] = useState<Tab>('schedule');
  const [expanded, setExpanded] = useState(false);
  const history = useRecurringTransactionHistory(props.recurring);
  const similar = useSimilarTransactions(props.recurring.transactionClusterId);

  // On schedule is everything monetr matched to this recurring transaction, that's all loaded at once so only the first
  // few show until you ask for the rest. All is the whole similar transactions group, which is paged so that one loads
  // more instead.
  let transactions = similar.data ?? [];
  let isLoading = similar.isLoading;
  if (tab === 'schedule') {
    transactions = history.transactions;
    isLoading = history.isLoading;
  }
  let visible = transactions;
  if (tab === 'schedule' && !expanded) {
    visible = transactions.slice(0, 5);
  }
  const hidden = transactions.length - visible.length;
  const canLoadMore = tab === 'all' && similar.hasNextPage;

  let heading = 'Deposits';
  if (props.recurring.direction === 'debit') {
    heading = 'Charges';
  }

  let showMore = 'Show More';
  if (similar.isFetchingNextPage) {
    showMore = 'Loading...';
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
            On Schedule {history.seen}
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
        {visible.map(item => (
          <RecurringChargeItem key={item.transactionId} recurring={props.recurring} transaction={item} />
        ))}
      </ul>

      {!isLoading && visible.length === 0 && !canLoadMore && (
        <Typography className={styles.empty} color='subtle' size='sm'>
          Nothing here...
        </Typography>
      )}
      {hidden > 0 && (
        <button className={styles.showMore} onClick={() => setExpanded(true)} type='button'>
          Show {hidden} More
        </button>
      )}
      {canLoadMore && (
        <button
          className={styles.showMore}
          disabled={similar.isFetchingNextPage}
          onClick={() => similar.fetchNextPage()}
          type='button'
        >
          {showMore}
        </button>
      )}
    </section>
  );
}
