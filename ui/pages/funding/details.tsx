import React from 'react';
import { useParams } from 'react-router-dom';
import { HeartBroken } from '@mui/icons-material';

import MSpan from 'components/MSpan';
import { useFundingSchedule } from 'hooks/fundingSchedules';

export default function FundingDetails(): JSX.Element {
  const { fundingId } = useParams();

  const funding = useFundingSchedule(fundingId && +fundingId);

  if (!fundingId) {
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

  if (!funding) {
    return null;
  }
  return null;
}
