import React, { Fragment } from 'react';
import { AccountBalance, AddOutlined, ArrowForward, Today, TodayOutlined } from '@mui/icons-material';

import FundingItem from './new/FundingItem';
import { showNewFundingModal } from './new/NewFundingModal';

import { MBaseButton } from 'components/MButton';
import MSidebarToggle from 'components/MSidebarToggle';
import MSpan from 'components/MSpan';
import { useFundingSchedulesSink } from 'hooks/fundingSchedules';

export default function FundingNew(): JSX.Element {
  const { isError: fundingIsError, isLoading: fundingIsLoading, data: funding } = useFundingSchedulesSink();

  if (fundingIsLoading) {
    return <MSpan>Loading...</MSpan>;
  }

  if (fundingIsError) {
    return <MSpan>Error...</MSpan>;
  }

  function EmptyState(): JSX.Element {
    return (
      <div className='w-full h-full flex justify-center items-center'>
        <div className='flex flex-col gap-2 items-center max-w-md'>
          <div className='w-full flex justify-center space-x-4'>
            <Today className='h-full text-5xl dark:text-dark-monetr-content-muted' />
            <ArrowForward className='h-full text-5xl dark:text-dark-monetr-content-muted' />
            <AccountBalance className='h-full text-5xl dark:text-dark-monetr-content-muted' />
          </div>
          <MSpan size='xl' variant='light' className='text-center'>
            You don't have any funding schedules yet...
          </MSpan>
          <MSpan size='lg' variant='light' className='text-center'>
            Funding schedules tell monetr when to allocate funds towards your expenses and goals.
          </MSpan>
        </div>
      </div>
    );
  }

  function ListContent(): JSX.Element {
    if (funding.length === 0) {
      return <EmptyState />;
    }

    return (
      <ul className='w-full flex flex-col gap-2 py-2'>
        { funding
          ?.sort((a, b) => a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1)
          .map(item => (<FundingItem funding={ item } key={ item.fundingScheduleId } />)) }
      </ul>
    );
  }

  return (
    <Fragment>
      <div className='w-full h-12 flex items-center px-4 gap-4 justify-between'>
        <div className='flex items-center gap-4'>
          <MSidebarToggle />
          <span className='text-2xl dark:text-dark-monetr-content-emphasis font-bold flex gap-2 items-center'>
            <TodayOutlined />
            Funding Schedules
          </span>
        </div>
        <MBaseButton color='primary' className='gap-1 py-1 px-2' onClick={ showNewFundingModal }>
          <AddOutlined />
          New Funding Schedule
        </MBaseButton>
      </div>
      <div className='w-full h-full overflow-y-auto min-w-0'>
        <ListContent />
      </div>
    </Fragment>
  );
}

