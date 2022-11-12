import React from 'react';
import { Button, ButtonGroup, Divider,  List } from '@mui/material';

import { showCreateExpenseDialog } from 'components/Expenses/CreateExpenseDialog';
import { showCreateFundingScheduleDialog } from 'components/FundingSchedules/CreateFundingScheduleDialog';
import FundingScheduleListItem from 'components/FundingSchedules/FundingScheduleListItem';
import { useFundingSchedules } from 'hooks/fundingSchedules';

export default function FundingScheduleList(): JSX.Element {
  const fundingSchedules = useFundingSchedules();

  return (
    <div className="w-full funding-schedule-list">
      <div className="w-full p-5">
        <ButtonGroup color="primary" className="w-full">
          <Button
            variant="outlined"
            className="w-full"
            color="primary"
            onClick={ showCreateFundingScheduleDialog }
          >
            New Funding Schedule
          </Button>
          <Button
            variant="outlined"
            className="w-full"
            color="primary"
            onClick={ showCreateExpenseDialog }
          >
            New Expense
          </Button>
        </ButtonGroup>
      </div>
      <Divider />
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
    </div>
  );
}
