import { Fragment } from 'react';
import { format } from 'date-fns';
import { Layers, Repeat } from 'lucide-react';

import { useRecurringTransactionHistory } from '@monetr/interface/hooks/useRecurringTransactionHistory';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { useTransactionClusterTransactions } from '@monetr/interface/hooks/useTransactionClusterTransactions';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';

import styles from './RecurringClusterCard.module.scss';

export interface RecurringClusterCardProps {
  recurring: TransactionRecurring;
  name: string;
}

export default function RecurringClusterCard({ recurring, name }: RecurringClusterCardProps): React.JSX.Element {
  const { inTimezone } = useTimezone();
  const { seen } = useRecurringTransactionHistory(recurring);
  const { data: transactions } = useTransactionClusterTransactions(recurring.transactionClusterId);

  // Count up each different way the charges in the group showed up on the bank statement, most common first. This is
  // only the most recent 100 but thats years worth for most things that repeat.
  const memoCounts: Record<string, number> = {};
  for (const transaction of transactions ?? []) {
    if (transaction.originalName) {
      memoCounts[transaction.originalName] = (memoCounts[transaction.originalName] ?? 0) + 1;
    }
  }
  const memos = Object.entries(memoCounts).sort((a, b) => b[1] - a[1]);

  return (
    <section className={styles.root}>
      <span className={styles.eyebrow}>
        <Layers />
        Similar transactions group
      </span>
      <span className={styles.title}>
        {name} <span>· since {format(inTimezone(recurring.first), 'MMM yyyy')}</span>
      </span>
      <div className={styles.chips}>
        <span className={styles.chip}>
          <Repeat />
          {seen} on this schedule
        </span>
      </div>
      {memos.length > 0 && (
        <div className={styles.memos}>
          <span className={styles.memosLabel}>Grouped because they show up as</span>
          <div className={styles.memoList}>
            {memos.map(([memo, count]) => (
              <Fragment key={memo}>
                <span className={styles.memo} title={memo}>
                  {memo}
                </span>
                <span className={styles.memoCount}>{count === 1 ? '1 charge' : `${count} charges`}</span>
              </Fragment>
            ))}
          </div>
        </div>
      )}
      <span className={styles.note}>
        Anything else from {name || 'this group'} that isn&apos;t on this schedule shows up as one-off below.
      </span>
    </section>
  );
}
