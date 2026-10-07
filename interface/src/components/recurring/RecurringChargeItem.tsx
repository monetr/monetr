import { ChevronRight } from 'lucide-react';
import { Link } from 'wouter';

import { flexVariants } from '@monetr/interface/components/Flex';
import { Item, ItemContent } from '@monetr/interface/components/Item';
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

export default function RecurringChargeItem({
  recurring,
  transaction,
}: RecurringChargeItemProps): React.JSX.Element | null {
  const { inTimezone } = useTimezone();
  const { data: locale } = useLocale();

  if (!locale) {
    return null;
  }

  // Anything in the group that monetr didn't match to this schedule is a one-off as far as this page cares.
  const isOneOff = transaction.transactionRecurringId !== recurring.transactionRecurringId;

  return (
    <Item>
      <Link
        className={flexVariants({ orientation: 'row', align: 'center' })}
        to={`/bank/${transaction.bankAccountId}/transactions/${transaction.transactionId}/details`}
      >
        <TransactionMerchantIcon
          name={transaction.getName()}
          pending={transaction.isPending}
          transactionRecurringId={isOneOff ? null : recurring.transactionRecurringId}
        />
        {/* Every charge here is the same merchant, so the date goes first and the memo is what tells them apart. */}
        <ItemContent align='default' flex='shrink' gap='none' justify='start' orientation='column' shrink='default'>
          <span className={styles.date}>
            {formatDate(transaction.date, inTimezone, locale, DateLength.Full)}
            {isOneOff && <span className={styles.oneOffChip}>One-off</span>}
          </span>
          <span className={styles.memo} title={transaction.originalName}>
            {transaction.originalName}
          </span>
        </ItemContent>
        <ItemContent align='center' flex='grow' justify='end' shrink='none' width='fit'>
          <TransactionAmount transaction={transaction} />
          <Typography>
            <ChevronRight />
          </Typography>
        </ItemContent>
      </Link>
    </Item>
  );
}
