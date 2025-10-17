import React, { Fragment } from 'react';
import { CalendarSync, Plus } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import FundingItem from '@monetr/interface/components/funding/FundingItem';
import MSpan from '@monetr/interface/components/MSpan';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import { useFundingSchedules } from '@monetr/interface/hooks/useFundingSchedules';
import { showNewFundingModal } from '@monetr/interface/modals/NewFundingModal';

export default function Funding(): JSX.Element {
  const { isError: fundingIsError, isLoading: fundingIsLoading, data: funding } = useFundingSchedules();

  if (fundingIsLoading) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <MSpan className='text-5xl'>One moment...</MSpan>
      </div>
    );
  }

  if (fundingIsError) {
    return <MSpan>Error...</MSpan>;
  }

  function ListContent(): JSX.Element {
    if (funding.length === 0) {
      return <EmptyState />;
    }

    return (
      <ul className='w-full flex flex-col gap-2 py-2 pb-16'>
        {funding
          ?.sort((a, b) => (a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1))
          .map(item => (
            <FundingItem funding={item} key={item.fundingScheduleId} />
          ))}
      </ul>
    );
  }

  return (
    <Fragment>
      <MTopNavigation icon={CalendarSync} title='Funding Schedules'>
        <Button variant='primary' onClick={showNewFundingModal}>
          <Plus />
          New Funding Schedule
        </Button>
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
          <CalendarSync className='dark:text-dark-monetr-content-muted h-12 w-12' />
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
