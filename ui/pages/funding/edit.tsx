import React from 'react';
import { useParams } from 'react-router-dom';

export default function FundingEditPage(): JSX.Element {
  const params = useParams();
  const fundingScheduleId = params['fundingScheduleId'];
  return (
    <div className='minus-nav'>
      <div className='w-full view-area bg-white'>
        <h1>Edit funding schedule: { fundingScheduleId }</h1>
      </div>
    </div>
  )
}
