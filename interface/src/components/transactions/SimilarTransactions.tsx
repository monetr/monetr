import MSpan from '@monetr/interface/components/MSpan';
import SimilarTransactionItem from '@monetr/interface/components/transactions/SimilarTransactionItem';
import { useSimilarTransactions } from '@monetr/interface/hooks/useSimilarTransactions';
import type Transaction from '@monetr/interface/models/Transaction';

export interface SimilarTransactionsProps {
  transaction: Transaction;
}

export default function SimilarTransactions(props: SimilarTransactionsProps): JSX.Element {
  const {
    data: similarData,
    isLoading,
    isError,
  } = useSimilarTransactions(props.transaction.transactionClusterMember?.transactionClusterId);

  if (isLoading) {
    return null;
  }

  if (isError) {
    return null;
  }

  if (!similarData) {
    return null;
  }

  // TODO Doesn't exclude the current transaction.
  const items = similarData.map(item => <SimilarTransactionItem key={item.transactionId} transaction={item} />);

  return (
    <div className='w-full flex flex-col gap-2'>
      <MSpan className='pl-4' size='xl' weight='semibold'>
        Similar Transactions
      </MSpan>
      <ul className='w-full flex gap-2 flex-col'>{items}</ul>
    </div>
  );
}
