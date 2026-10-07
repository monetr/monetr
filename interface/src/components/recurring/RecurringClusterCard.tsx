import { Fragment } from 'react';
import { format } from 'date-fns';
import { Layers, Repeat } from 'lucide-react';

import RecurringMemo from '@monetr/interface/components/recurring/RecurringMemo';
import { useRecurringTransactionHistory } from '@monetr/interface/hooks/useRecurringTransactionHistory';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { useTransactionClusterTransactions } from '@monetr/interface/hooks/useTransactionClusterTransactions';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';

import styles from './RecurringClusterCard.module.scss';

export interface RecurringClusterCardProps {
  recurring: TransactionRecurring;
  name: string;
}

export default function RecurringClusterCard(props: RecurringClusterCardProps): React.JSX.Element {
  const { inTimezone } = useTimezone();
  const { seen } = useRecurringTransactionHistory(props.recurring);
  const { data: transactions } = useTransactionClusterTransactions(props.recurring.transactionClusterId);

  // Count up each different way the charges in the group showed up on the bank statement, most common first. This is
  // only the most recent 100 but thats years worth for most things that repeat.
  const memoCounts = (transactions ?? []).reduce<Record<string, number>>((counts, item) => {
    if (item.originalName) {
      counts[item.originalName] = (counts[item.originalName] ?? 0) + 1;
    }
    return counts;
  }, {});
  const memos = Object.entries(memoCounts).sort((a, b) => b[1] - a[1]);

  return (
    <section className={styles.root}>
      <span className={styles.eyebrow}>
        <Layers />
        Similar Transactions Group
      </span>
      <span className={styles.title}>
        {props.name} <span>&middot; since {format(inTimezone(props.recurring.first), 'MMM yyyy')}</span>
      </span>
      <div className={styles.chips}>
        <span className={styles.chip} data-testid='recurring-on-schedule'>
          <Repeat />
          {seen} on this schedule
        </span>
      </div>
      {memos.length > 0 && (
        <div className={styles.memos}>
          <span className={styles.memosLabel}>Grouped Because They Show Up As</span>
          <div className={styles.memoList} data-testid='recurring-memos'>
            {memos.map(([memo, count]) => (
              <Fragment key={memo}>
                <RecurringMemo className={styles.memo} memo={memo} />
                <span className={styles.memoCount}>
                  {count === 1 && '1 charge'}
                  {count !== 1 && `${count} charges`}
                </span>
              </Fragment>
            ))}
          </div>
        </div>
      )}
      <span className={styles.note}>
        Anything else from {props.name || 'this group'} that isn&apos;t on this schedule shows up as one-off below.
      </span>
    </section>
  );
}
