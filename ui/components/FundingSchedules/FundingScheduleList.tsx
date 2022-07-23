import React, { useState } from 'react';
import { Button, ButtonGroup, Divider,  List } from '@mui/material';

import NewExpenseDialog from 'components/Expenses/NewExpenseDialog';
import FundingScheduleListItem from 'components/FundingSchedules/FundingScheduleListItem';
import NewFundingScheduleDialog from 'components/FundingSchedules/NewFundingScheduleDialog';
import { useFundingSchedules } from 'hooks/fundingSchedules';

export default function FundingScheduleList(): JSX.Element {
  enum DialogOpen {
    NewFundingSchedule,
    NewExpense,
  }

  const [currentDialog, setOpenDialog] = useState<DialogOpen | null>(null);
  const fundingSchedules = useFundingSchedules();

  function openDialog(dialog: DialogOpen) {
    setOpenDialog(dialog);
  }

  function Dialog(): JSX.Element {
    function closeDialog() {
      setOpenDialog(null);
    }

    switch (currentDialog) {
      case DialogOpen.NewFundingSchedule:
        return <NewFundingScheduleDialog onClose={ closeDialog } isOpen />;
      case DialogOpen.NewExpense:
        return <NewExpenseDialog onClose={ closeDialog } isOpen />;
      default:
        return null;
    }
  }

  return (
    <div className="w-full funding-schedule-list">
      <Dialog />
      <div className="w-full p-5">
        <ButtonGroup color="primary" className="w-full">
          <Button variant="outlined" className="w-full" color="primary"
            onClick={ () => openDialog(DialogOpen.NewFundingSchedule) }>
            New Funding Schedule
          </Button>
          <Button variant="outlined" className="w-full" color="primary"
            onClick={ () => openDialog(DialogOpen.NewExpense) }>
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
