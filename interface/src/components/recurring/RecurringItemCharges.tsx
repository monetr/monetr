import { format, isThisYear } from 'date-fns';
import { Layers } from 'lucide-react';
import { Link } from 'wouter';

import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useRecurringTransactionHistory } from '@monetr/interface/hooks/useRecurringTransactionHistory';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import { AmountType } from '@monetr/interface/util/amounts';

import styles from './RecurringItemCharges.module.scss';

export interface RecurringItemChargesProps {
  recurring: TransactionRecurring;
  name: string;
  detailsPath: string;
}

// The peek you get when you expand a recurring item. Shows the last few charges with what they actually looked like on
// the bank statement. Its only rendered once the item is expanded so the transactions aren't fetched until then.
export default function RecurringItemCharges(props: RecurringItemChargesProps): React.JSX.Element | null {
  const { data: locale } = useLocaleCurrency();
  const { inTimezone } = useTimezone();
  const { transactions, seen, isLoading } = useRecurringTransactionHistory(props.recurring);

  if (!locale) {
    return null;
  }

  return (
    <div className={styles.root}>
      <div className={styles.header}>
        <Layers />
        <span>Similar transactions group</span>
        <strong>{props.name}</strong>
        <span>
          · {seen === 1 ? '1 charge' : `${seen} charges`} since {format(inTimezone(props.recurring.first), 'MMM yyyy')}
        </span>
        <Link className={styles.detailsLink} to={props.detailsPath}>
          View details
        </Link>
      </div>
      {isLoading && <span className={styles.loading}>Loading...</span>}
      {!isLoading && (
        <div className={styles.list}>
          {transactions.slice(0, 3).map(transaction => (
            <div className={styles.charge} key={transaction.transactionId}>
              <span className={styles.date}>
                {isThisYear(transaction.date)
                  ? format(inTimezone(transaction.date), 'MMM d')
                  : format(inTimezone(transaction.date), 'MMM d, yyyy')}
              </span>
              <span className={styles.memo} title={transaction.originalName}>
                {transaction.originalName}
              </span>
              <span className={styles.amount} data-addition={String(transaction.getIsAddition())}>
                {locale.formatAmount(Math.abs(transaction.amount), AmountType.Stored)}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
