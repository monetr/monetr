import React from 'react';
import { AccountBalance, Add, ArrowForward, Today } from '@mui/icons-material';
import { Button, CircularProgress, Fab, List, Typography } from '@mui/material';

import { showCreateFundingScheduleDialog } from 'components/FundingSchedules/CreateFundingScheduleDialog';
import FundingScheduleListItem from 'components/FundingSchedules/FundingScheduleListItem';
import { useFundingSchedulesSink } from 'hooks/fundingSchedules';

export default function FundingSchedulesView(): JSX.Element {
  const { isLoading, result: fundingSchedules } = useFundingSchedulesSink();

  if (isLoading) {
    return (
      <div className="h-full w-full bg-primary">
        <div className="view-inner h-full flex justify-center items-center">
          <div className="grid grid-cols-1 grid-rows-2 grid-flow-col gap-2">
            <CircularProgress />
          </div>
        </div>
      </div>
    );
  }

  function EmptyState(): JSX.Element {
    return (
      <div className="h-full w-full bg-primary">
        <div className="view-inner h-full flex justify-center items-center">
          <div className="grid grid-cols-1 grid-rows-3 grid-flow-col gap-2">
            <div className="w-full flex justify-center space-x-4">
              <Today className='h-full text-5xl opacity-50' />
              <ArrowForward className='h-full text-5xl opacity-50' />
              <AccountBalance className='h-full text-5xl opacity-50' />
            </div>
            <Typography
              className="opacity-50"
              variant="h3"
            >
              You don't have any funding schedules yet...
            </Typography>
            <Button
              onClick={ showCreateFundingScheduleDialog }
              color="primary"
            >
              <Typography
                variant="h6"
              >
                Create A Funding Schedule
              </Typography>
            </Button>
          </div>
        </div>
      </div>
    );
  }

  if (fundingSchedules.size === 0) {
    return <EmptyState />;
  }

  return (
    <div className="minus-nav">
      <div className="w-full view-area bg-white">
        <List className="w-full pt-0" dense>
          {
            Array.from(fundingSchedules.values())
              .map(schedule => (
                <FundingScheduleListItem
                  key={ schedule.fundingScheduleId }
                  fundingScheduleId={ schedule.fundingScheduleId }
                />
              ))
          }
        </List>
        <Fab
          color="primary"
          aria-label="add"
          className="absolute z-50 bottom-5 right-5"
          onClick={ showCreateFundingScheduleDialog }
        >
          <Add />
        </Fab>
      </div>
    </div>
  );
}
