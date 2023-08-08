import React, { Fragment } from 'react';
import { AddOutlined, TodayOutlined } from '@mui/icons-material';

import FundingItem from './new/FundingItem';
import { showNewFundingModal } from './new/NewFundingModal';

import { MBaseButton } from 'components/MButton';
import MSidebarToggle from 'components/MSidebarToggle';
import MSpan from 'components/MSpan';
import { useFundingSchedulesSink } from 'hooks/fundingSchedules';

export default function FundingNew(): JSX.Element {
  const { isError: fundingIsError, isLoading: fundingIsLoading, result: funding } = useFundingSchedulesSink();

  if (fundingIsLoading) {
    return <MSpan>Loading...</MSpan>;
  }

  if (fundingIsError) {
    return <MSpan>Error...</MSpan>;
  }

  return (
    <Fragment>
      <div className='w-full h-12 flex items-center px-4 gap-4 justify-between'>
        <div className='flex items-center gap-4'>
          <MSidebarToggle />
          <span className='text-2xl dark:text-dark-monetr-content-emphasis font-bold flex gap-2 items-center'>
            <TodayOutlined />
            Expenses
          </span>
        </div>
        <MBaseButton color='primary' className='gap-1 py-1 px-2' onClick={ showNewFundingModal }>
          <AddOutlined />
          New Funding Schedule
        </MBaseButton>
      </div>
      <div className='w-full h-full overflow-y-auto min-w-0'>
        <ul className='w-full flex flex-col gap-2 py-2'>
          { Array.from(funding.values())
            ?.sort((a, b) => a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1)
            .map(item => (<FundingItem funding={ item } key={ item.fundingScheduleId } />)) }
        </ul>
      </div>
    </Fragment>
  );
}

