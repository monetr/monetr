import { Link } from 'wouter';

import { flexVariants } from '@monetr/interface/components/Flex';
import { Item, ItemContent } from '@monetr/interface/components/Item';
import RecurringMemo from '@monetr/interface/components/recurring/RecurringMemo';
import TransactionAmount from '@monetr/interface/components/transactions/TransactionAmount';
import TransactionMerchantIcon from '@monetr/interface/components/transactions/TransactionMerchantIcon';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import { useSpending } from '@monetr/interface/hooks/useSpending';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import type Transaction from '@monetr/interface/models/Transaction';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import { DateLength, formatDate } from '@monetr/interface/util/formatDate';

import styles from './RecurringChargeItem.module.scss';

export interface RecurringChargeItemProps {
  recurring: TransactionRecurring;
  transaction: Transaction;
}

export default function RecurringChargeItem(props: RecurringChargeItemProps): React.JSX.Element | null {
  const { inTimezone } = useTimezone();
  const { data: locale } = useLocale();
  const { data: spending } = useSpending(props.transaction.spendingId);

  if (!locale) {
    return null;
  }

  // Anything in the group that monetr didn't match to this schedule is a one-off as far as this page cares
  const isOneOff = props.transaction.transactionRecurringId !== props.recurring.transactionRecurringId;

  return (
    <Item>
      <Link
        className={flexVariants({ orientation: 'row', align: 'center' })}
        to={`/bank/${props.transaction.bankAccountId}/transactions/${props.transaction.transactionId}/details`}
      >
        <TransactionMerchantIcon name={props.transaction.getName()} pending={props.transaction.isPending} />
        {/* Every charge here is the same merchant, so the date goes first and the memo is what tells them apart */}
        <ItemContent align='default' flex='shrink' gap='none' justify='start' orientation='column' shrink='default'>
          <span className={styles.date}>
            {formatDate(props.transaction.date, inTimezone, locale, DateLength.Full)}
            {isOneOff && (
              <span className={styles.oneOffChip} data-testid='recurring-charge-one-off'>
                One-Off
              </span>
            )}
          </span>
          <RecurringMemo memo={props.transaction.originalName} />
        </ItemContent>
        <ItemContent align='center' flex='grow' justify='end' shrink='none' width='fit'>
          {!props.transaction.getIsAddition() && (
            <span className={styles.spentFrom}>Spent from {spending?.name || 'Free-To-Use'}</span>
          )}
          <TransactionAmount transaction={props.transaction} />
        </ItemContent>
      </Link>
    </Item>
  );
}
