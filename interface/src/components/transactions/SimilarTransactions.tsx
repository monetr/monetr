import Typography from '@monetr/interface/components/Typography';
import SimilarTransactionItem from '@monetr/interface/components/transactions/SimilarTransactionItem';
import { useSimilarTransactions } from '@monetr/interface/hooks/useSimilarTransactions';
import type Transaction from '@monetr/interface/models/Transaction';

import styles from './SimilarTransactions.module.scss';

export interface SimilarTransactionsProps {
  transaction: Transaction;
}

export default function SimilarTransactions(props: SimilarTransactionsProps): React.JSX.Element {
  const { data: similarData, isLoading, isError } = useSimilarTransactions(props.transaction);

  if (isLoading) {
    return null;
  }

  if (isError) {
    return null;
  }

  if (similarData?.members?.length === 0) {
    return null;
  }

  const maxNumberOfSimilarTransactions = 10;
  const items = similarData.members
    .filter(item => item !== props.transaction.transactionId)
    .slice(0, Math.min(maxNumberOfSimilarTransactions, similarData.members.length) - 1)
    .map(item => <SimilarTransactionItem key={item} transactionId={item} />);

  return (
    <div className={styles.root}>
      <Typography className={styles.heading} size='xl' weight='semibold'>
        Similar Transactions
      </Typography>
      <ul className={styles.list}>{items}</ul>
    </div>
  );
}
