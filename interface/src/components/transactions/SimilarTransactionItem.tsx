import React from 'react';
import { format, isThisYear } from 'date-fns';

import { useTransaction } from '@monetr/interface/hooks/transactions';
import TransactionMerchantIcon from '@monetr/interface/pages/new/TransactionMerchantIcon';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface SimilarTransactionItemProps {
  transactionId: number;
}

export default function SimilarTransactionItem(props: SimilarTransactionItemProps): JSX.Element {
  const { data: transaction, isLoading, isError } = useTransaction(props.transactionId);
  if (isLoading) {
    return null;
  }

  if (isError) {
    return null;
  }

  const amountClassnames = mergeTailwind(
    {
      'dark:text-dark-monetr-green': transaction.getIsAddition(),
      'dark:text-dark-monetr-red': !transaction.getIsAddition(),
    },
    'text-end',
    'font-semibold',
  );

  const dateString =  isThisYear(transaction.date) ?
    format(transaction.date, 'MMMM do') :
    format(transaction.date, 'MMMM do, yyyy');

  return (
    <li className='group relative w-full px-1 md:px-2'>
      <div className='group flex h-full gap-1 rounded-lg px-2 py-1 group-hover:bg-zinc-600 md:gap-4'>
        <div className='flex w-full min-w-0 flex-1 flex-row items-center gap-4 md:w-1/2'>
          <TransactionMerchantIcon name={ transaction.getName() } pending={ transaction.isPending } />
          <div className='flex min-w-0 flex-col overflow-hidden'>
            <span className='w-full min-w-0 truncate text-base font-semibold dark:text-dark-monetr-content-emphasis'>
              {transaction.getName()}
            </span>
            <span className='w-full min-w-0 truncate text-sm font-medium dark:text-dark-monetr-content'>
              { dateString }
            </span>
          </div>
        </div>
        <div className='flex shrink-0 items-center justify-end gap-2 md:min-w-[8em]'>
          <span className={ amountClassnames }>
            {transaction.getAmountString()}
          </span>
        </div>
      </div>
    </li>
  );
}
