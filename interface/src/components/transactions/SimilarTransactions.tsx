
import React from 'react';

import MSpan from '@monetr/interface/components/MSpan';
import SimilarTransactionItem from '@monetr/interface/components/transactions/SimilarTransactionItem';
import { useSimilarTransactions } from '@monetr/interface/hooks/useSimilarTransactions';
import Transaction from '@monetr/interface/models/Transaction';

export interface SimilarTransactionsProps {
  transaction: Transaction;
}

export default function SimilarTransactions(props: SimilarTransactionsProps): JSX.Element {
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
    .map(item => <SimilarTransactionItem key={ item } transactionId={ item } />);

  return (
    <div className='w-full flex flex-col gap-2'>
      <MSpan size='xl' weight='semibold' className='pl-4'>
        Similar Transactions
      </MSpan>
      <ul className='w-full flex gap-2 flex-col'>
        { items }
      </ul>
    </div>
  );
}
