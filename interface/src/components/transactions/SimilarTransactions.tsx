import { Clock } from 'lucide-react';

import { Item, ItemContent } from '@monetr/interface/components/Item';
import Typography from '@monetr/interface/components/Typography';
import SimilarTransactionItem from '@monetr/interface/components/transactions/SimilarTransactionItem';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useRecurringTransaction } from '@monetr/interface/hooks/useRecurringTransaction';
import { useSimilarTransactions } from '@monetr/interface/hooks/useSimilarTransactions';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import type Transaction from '@monetr/interface/models/Transaction';
import { AmountType } from '@monetr/interface/util/amounts';
import { DateLength, formatDate } from '@monetr/interface/util/formatDate';

import styles from './SimilarTransactions.module.scss';

export interface SimilarTransactionsProps {
  transaction: Transaction;
}

export default function SimilarTransactions(props: SimilarTransactionsProps): React.JSX.Element | null {
  const {
    data: similarData,
    isLoading,
    isError,
  } = useSimilarTransactions(props.transaction.transactionClusterId ?? undefined);

  if (isLoading) {
    return null;
  }

  if (isError) {
    return null;
  }

  if (!similarData || similarData.length === 0) {
    return null;
  }

  const items = similarData.map(item => (
    <SimilarTransactionItem
      current={item.transactionId === props.transaction.transactionId}
      key={item.transactionId}
      transaction={item}
    />
  ));

  return (
    <div className={styles.root}>
      <Typography className={styles.heading} size='xl' weight='semibold'>
        Similar Transactions
      </Typography>
      <ul className={styles.list}>
        <ExpectedTransactionItem transaction={props.transaction} />
        {items}
      </ul>
    </div>
  );
}

// ExpectedTransactionItem shows the next transaction we expect if this one is recurring, so you can see when the next one
// should show up right next to the ones that already did
function ExpectedTransactionItem({ transaction }: SimilarTransactionsProps): React.JSX.Element | null {
  const { data: recurring } = useRecurringTransaction(transaction.transactionRecurringId);
  const { inTimezone } = useTimezone();
  const { data: locale } = useLocale();
  const { data: localeCurrency } = useLocaleCurrency();

  if (!recurring || recurring.ended || !locale || !localeCurrency) {
    return null;
  }

  return (
    <Item className={styles.expected} data-testid='similar-transactions-expected'>
      <div className={styles.expectedIcon}>
        <Clock />
      </div>
      <ItemContent align='default' flex='shrink' gap='none' justify='start' orientation='column' shrink='default'>
        <Typography component='p' ellipsis size='md' weight='semibold'>
          {transaction.getName()}
        </Typography>
        <Typography color='subtle' component='p' ellipsis size='sm' weight='medium'>
          Expected {formatDate(recurring.next, inTimezone, locale, DateLength.Long)}
        </Typography>
      </ItemContent>
      <ItemContent align='center' flex='grow' justify='end' shrink='none' width='fit'>
        <Typography color='subtle' weight='semibold'>
          ~{localeCurrency.formatAmount(Math.abs(recurring.lastAmount), AmountType.Stored)}
        </Typography>
      </ItemContent>
    </Item>
  );
}
