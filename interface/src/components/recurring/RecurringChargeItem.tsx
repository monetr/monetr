import { ChevronRight } from 'lucide-react';
import { Link } from 'wouter';

import { flexVariants } from '@monetr/interface/components/Flex';
import { Item, ItemContent } from '@monetr/interface/components/Item';
import RecurringMemo from '@monetr/interface/components/recurring/RecurringMemo';
import Typography from '@monetr/interface/components/Typography';
import TransactionAmount from '@monetr/interface/components/transactions/TransactionAmount';
import TransactionMerchantIcon from '@monetr/interface/components/transactions/TransactionMerchantIcon';
import { useLocale } from '@monetr/interface/hooks/useLocale';
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

  if (!locale) {
    return null;
  }

  // Anything in the group that monetr didn't match to this schedule is a one-off as far as this page cares
  const isOneOff = props.transaction.transactionRecurringId !== props.recurring.transactionRecurringId;

  let recurringId = null;
  if (!isOneOff) {
    recurringId = props.recurring.transactionRecurringId;
  }

  return (
    <Item>
      <Link
        className={flexVariants({ orientation: 'row', align: 'center' })}
        to={`/bank/${props.transaction.bankAccountId}/transactions/${props.transaction.transactionId}/details`}
      >
        <TransactionMerchantIcon
          name={props.transaction.getName()}
          pending={props.transaction.isPending}
          transactionRecurringId={recurringId}
        />
        {/* Every charge here is the same merchant, so the date goes first and the memo is what tells them apart */}
        <ItemContent align='default' flex='shrink' gap='none' justify='start' orientation='column' shrink='default'>
          <span className={styles.date}>
            {formatDate(props.transaction.date, inTimezone, locale, DateLength.Full)}
            {isOneOff && <span className={styles.oneOffChip}>One-Off</span>}
          </span>
          <RecurringMemo memo={props.transaction.originalName} />
        </ItemContent>
        <ItemContent align='center' flex='grow' justify='end' shrink='none' width='fit'>
          <TransactionAmount transaction={props.transaction} />
          <Typography>
            <ChevronRight />
          </Typography>
        </ItemContent>
      </Link>
    </Item>
  );
}
