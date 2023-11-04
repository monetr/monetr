import React, { Fragment } from 'react';
import { AccountBalance, AddOutlined, ArrowForward, Today, TodayOutlined } from '@mui/icons-material';

import { MBaseButton } from '@monetr/interface/components/MButton';
import MSpan from '@monetr/interface/components/MSpan';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import { useFundingSchedulesSink } from '@monetr/interface/hooks/fundingSchedules';
import { showNewFundingModal } from '@monetr/interface/modals/NewFundingModal';

import FundingItem from './new/FundingItem';

export default function Funding(): JSX.Element {
  const { isError: fundingIsError, isLoading: fundingIsLoading, data: funding } = useFundingSchedulesSink();

  if (fundingIsLoading) {
    return <MSpan>Loading...</MSpan>;
  }

  if (fundingIsError) {
    return <MSpan>Error...</MSpan>;
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
      <MTopNavigation
        icon={ TodayOutlined }
        title='Funding Schedules'
      >
        <MBaseButton color='primary' className='gap-1 py-1 px-2' onClick={ showNewFundingModal }>
          <AddOutlined />
          New Funding Schedule
        </MBaseButton>
      </MTopNavigation>
      <div className='w-full h-full overflow-y-auto min-w-0'>
        <ListContent />
      </div>
    </Fragment>
  );
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
        <MSpan size='xl' color='subtle' className='text-center'>
            You don't have any funding schedules yet...
        </MSpan>
        <MSpan size='lg' color='subtle' className='text-center'>
            Funding schedules tell monetr when to allocate funds towards your expenses and goals.
        </MSpan>
      </div>
    </div>
  );
}

