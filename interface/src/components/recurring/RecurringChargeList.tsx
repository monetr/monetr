import { useState } from 'react';
import { ChevronRight, Clock } from 'lucide-react';

import { flexVariants } from '@monetr/interface/components/Flex';
import { Item, ItemContent } from '@monetr/interface/components/Item';
import RecurringChargeItem from '@monetr/interface/components/recurring/RecurringChargeItem';
import Typography from '@monetr/interface/components/Typography';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useRecurringTransactionHistory } from '@monetr/interface/hooks/useRecurringTransactionHistory';
import { useSimilarTransactions } from '@monetr/interface/hooks/useSimilarTransactions';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import { AmountType } from '@monetr/interface/util/amounts';
import { DateLength, formatDate } from '@monetr/interface/util/formatDate';

import styles from './RecurringChargeList.module.scss';

type Tab = 'all' | 'schedule' | 'oneoff';

export interface RecurringChargeListProps {
  recurring: TransactionRecurring;
}

export default function RecurringChargeList({ recurring }: RecurringChargeListProps): React.JSX.Element | null {
  const { data: locale } = useLocaleCurrency();
  const { data: dateLocale } = useLocale();
  const { inTimezone } = useTimezone();
  const [tab, setTab] = useState<Tab>('all');
  const history = useRecurringTransactionHistory(recurring);
  const similar = useSimilarTransactions(recurring.transactionClusterId);

  if (!locale || !dateLocale) {
    return null;
  }

  // On schedule is everything monetr matched to this recurring transaction. The other two come from the whole similar
  // transactions group, which is paged so those can load more.
  let transactions = similar.data ?? [];
  if (tab === 'schedule') {
    transactions = history.transactions;
  } else if (tab === 'oneoff') {
    transactions = transactions.filter(item => item.transactionRecurringId !== recurring.transactionRecurringId);
  }
  const canLoadMore = tab !== 'schedule' && similar.hasNextPage;
  const isLoading = tab === 'schedule' ? history.isLoading : similar.isLoading;

  return (
    <section className={styles.root}>
      <div className={styles.header}>
        <Typography component='h3' size='xl' weight='semibold'>
          {recurring.direction === 'debit' ? 'Charges' : 'Deposits'}
        </Typography>
        <div className={styles.tabs} role='tablist'>
          <button
            aria-selected={tab === 'all'}
            className={styles.tab}
            onClick={() => setTab('all')}
            role='tab'
            type='button'
          >
            All
          </button>
          <button
            aria-selected={tab === 'schedule'}
            className={styles.tab}
            onClick={() => setTab('schedule')}
            role='tab'
            type='button'
          >
            On schedule {history.seen}
          </button>
          <button
            aria-selected={tab === 'oneoff'}
            className={styles.tab}
            onClick={() => setTab('oneoff')}
            role='tab'
            type='button'
          >
            One-off
          </button>
        </div>
      </div>

      <ul className={styles.list}>
        {/* The next one we expect, one-offs by definition don't have one. */}
        {tab !== 'oneoff' && !recurring.ended && (
          <Item className={styles.expected}>
            <div className={flexVariants({ orientation: 'row', align: 'center' })}>
              <div className={styles.expectedIcon}>
                <Clock />
              </div>
              <ItemContent
                align='default'
                flex='shrink'
                gap='none'
                justify='start'
                orientation='column'
                shrink='default'
              >
                <Typography component='p' ellipsis size='md' weight='semibold'>
                  {formatDate(recurring.next, inTimezone, dateLocale, DateLength.Full)}
                </Typography>
                <Typography color='subtle' component='p' ellipsis size='sm' weight='medium'>
                  Expected next
                </Typography>
              </ItemContent>
              <ItemContent align='center' flex='grow' justify='end' shrink='none' width='fit'>
                <Typography color='subtle' weight='semibold'>
                  ~{locale.formatAmount(Math.abs(recurring.lastAmount), AmountType.Stored)}
                </Typography>
                {/* Nothing to link to, but keep the arrow's space so the amount lines up with the real rows. */}
                <Typography aria-hidden className={styles.arrowSpacer}>
                  <ChevronRight />
                </Typography>
              </ItemContent>
            </div>
          </Item>
        )}
        {transactions.map(transaction => (
          <RecurringChargeItem key={transaction.transactionId} recurring={recurring} transaction={transaction} />
        ))}
      </ul>

      {!isLoading && transactions.length === 0 && !canLoadMore && (
        <Typography className={styles.empty} color='subtle' size='sm'>
          Nothing here.
        </Typography>
      )}
      {canLoadMore && (
        <button
          className={styles.showMore}
          disabled={similar.isFetchingNextPage}
          onClick={() => similar.fetchNextPage()}
          type='button'
        >
          {similar.isFetchingNextPage ? 'Loading...' : 'Show more'}
        </button>
      )}
    </section>
  );
}
