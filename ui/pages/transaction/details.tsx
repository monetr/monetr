import React from 'react';
import { useParams } from 'react-router-dom';
import { HeartBroken } from '@mui/icons-material';

import MSpan from 'components/MSpan';
import { useTransaction } from 'hooks/transactions';

export default function TransactionDetails(): JSX.Element {
  const { transactionId: id } = useParams();
  const transactionId = +id || null;

  const { result: transaction, isLoading, isError } = useTransaction(transactionId);
  if (!transactionId) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartBroken className='dark:text-dark-monetr-content h-24 w-24' />
        <MSpan className='text-5xl'>
          Something isn't right...
        </MSpan>
        <MSpan className='text-2xl'>
          There wasn't an expense specified...
        </MSpan>
      </div>
    );
  }
  if (isError) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartBroken className='dark:text-dark-monetr-content h-24 w-24' />
        <MSpan className='text-5xl'>
          Something isn't right...
        </MSpan>
        <MSpan className='text-2xl'>
          We weren't able to load details for the transaction specified...
        </MSpan>
      </div>
    );
  }

  return null;
}
