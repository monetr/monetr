import { format } from 'date-fns';
import { Layers } from 'lucide-react';
import { Link } from 'wouter';

import RecurringMemo from '@monetr/interface/components/recurring/RecurringMemo';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { useTransactionsForRecurring } from '@monetr/interface/hooks/useTransactionsForRecurring';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import { AmountType } from '@monetr/interface/util/amounts';
import { DateLength, formatDate } from '@monetr/interface/util/formatDate';

import styles from './RecurringItemCharges.module.scss';

export interface RecurringItemChargesProps {
  recurring: TransactionRecurring;
  name: string;
  detailsPath: string;
}

// The peek you get when you expand a recurring item, the last few charges with what they looked like on the bank
// statement. Its only rendered once the item is expanded so the transactions aren't fetched until then.
export default function RecurringItemCharges(props: RecurringItemChargesProps): React.JSX.Element | null {
  const { data: locale } = useLocaleCurrency();
  const { data: dateLocale } = useLocale();
  const { inTimezone } = useTimezone();
  // Only the last three show up here so that's all we ask for
  const { data: transactions, isLoading } = useTransactionsForRecurring(props.recurring.transactionRecurringId, 3);

  if (!locale || !dateLocale) {
    return null;
  }

  return (
    <div className={styles.root}>
      <div className={styles.header}>
        <Layers />
        <span>Similar Transactions Group</span>
        <strong>{props.name}</strong>
        <span>&middot; since {format(inTimezone(props.recurring.first), 'MMM yyyy')}</span>
        <Link className={styles.detailsLink} to={props.detailsPath}>
          View Details
        </Link>
      </div>
      {isLoading && <span className={styles.loading}>Loading...</span>}
      {!isLoading && (
        <div className={styles.list}>
          {(transactions ?? []).map(item => (
            <div className={styles.charge} key={item.transactionId}>
              <span className={styles.date}>{formatDate(item.date, inTimezone, dateLocale, DateLength.Medium)}</span>
              <RecurringMemo className={styles.memo} memo={item.originalName} />
              <span className={styles.amount} data-addition={String(item.getIsAddition())}>
                {locale.formatAmount(Math.abs(item.amount), AmountType.Stored)}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
