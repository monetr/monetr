import Typography from '@monetr/interface/components/Typography';
import SimilarTransactionItem from '@monetr/interface/components/transactions/SimilarTransactionItem';
import { useSimilarTransactions } from '@monetr/interface/hooks/useSimilarTransactions';
import type Transaction from '@monetr/interface/models/Transaction';

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
      <ul className={styles.list}>{items}</ul>
    </div>
  );
}
