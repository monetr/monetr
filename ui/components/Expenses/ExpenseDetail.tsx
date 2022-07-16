import React, { Fragment, useState } from 'react';
import {
  AccountBalance,
  ArrowBack,
  ArrowForward,
  ChevronRight,
  DeleteOutline,
  Event,
  SwapHoriz,
  TrackChanges,
} from '@mui/icons-material';
import { Button, Divider, LinearProgress, List, ListItem, ListItemIcon, Typography } from '@mui/material';

import EditSpendingAmountDialog from 'components/Expenses/EditExpenseAmountDialog';
import EditExpenseDueDateDialog from 'components/Expenses/EditExpenseDueDateDialog';
import FundingScheduleList from 'components/FundingSchedules/FundingScheduleList';
import TransferDialog from 'components/Spending/TransferDialog';
import { useFundingSchedule } from 'hooks/fundingSchedules';
import { useRemoveSpending, useSelectedExpense } from 'hooks/spending';

export default function ExpenseDetail(): JSX.Element {
  const removeSpending = useRemoveSpending();
  const expense = useSelectedExpense();
  const fundingSchedule = useFundingSchedule(expense?.fundingScheduleId);

  enum Dialog {
    TransferDialog,
    EditAmountDialog,
    EditDueDateDialog,
    EditFundingScheduleDialog,
  }

  const [dialogOpen, setDialogOpen] = useState<Dialog | null>(null);

  function openDialog(dialog: Dialog) {
    setDialogOpen(dialog);
  }

  function closeDialog() {
    setDialogOpen(null);
  }

  function Dialogs(): JSX.Element {
    switch (dialogOpen) {
      case Dialog.TransferDialog:
        return <TransferDialog initialToSpendingId={ expense.spendingId } isOpen onClose={ closeDialog } />;
      case Dialog.EditAmountDialog:
        return <EditSpendingAmountDialog spending={ expense } isOpen onClose={ closeDialog } />;
      case Dialog.EditDueDateDialog:
        return <EditExpenseDueDateDialog spending={ expense } isOpen onClose={ closeDialog } />;
      case Dialog.EditFundingScheduleDialog:
        // TODO Implement.
        return null;
      default:
        return null;
    }
  }

  async function deleteExpense(): Promise<void> {
    if (!expense) {
      return Promise.resolve();
    }

    if (window.confirm(`Are you sure you want to delete expense: ${ expense.name }`)) {
      return removeSpending(expense.spendingId);
    }

    return Promise.resolve();
  }

  if (!expense) {
    return (<FundingScheduleList />);
  }

  return (
    <Fragment>
      <Dialogs />
      <div className="w-full pl-5 pr-5 pt-5 expense-detail">
        <div className="grid grid-cols-3 grid-rows-4 grid-flow-col gap-1 w-auto">
          <div className="col-span-2">
            <Typography
              variant="h5"
            >
              { expense.name }
            </Typography>
          </div>
          <div className="col-span-2">
            <Typography
              variant="h6"
            >
              { expense.getCurrentAmountString() } of { expense.getTargetAmountString() }
            </Typography>
          </div>
          <div className="col-span-3">
            <Typography>
              { expense.getNextOccurrenceString() } - { expense.description }
            </Typography>
          </div>
          <div className="col-span-3">
            <Typography>
              { expense.getNextContributionAmountString() }/{ fundingSchedule?.name }
            </Typography>
          </div>
          <div className="col-span-1 row-span-2">
            <LinearProgress
              className="mt-3"
              variant="determinate"
              value={ (expense.currentAmount / expense.targetAmount) * 100 }
            />
          </div>
        </div>

        <List dense>
          <Divider />
          <ListItem button dense onClick={ () => openDialog(Dialog.EditAmountDialog) }>
            <ListItemIcon>
              <AccountBalance />
            </ListItemIcon>
            <div className="grid grid-cols-3 grid-rows-2 grid-flow-col gap-1 w-full">
              <div className="col-span-3">
                <Typography>
                  Amount
                </Typography>
              </div>
              <div className="col-span-3 opacity-50">
                <Typography variant="body2">
                  { expense.getTargetAmountString() }
                </Typography>
              </div>
              <div className="col-span-1 row-span-2 flex justify-end">
                <ChevronRight className="align-middle h-full" />
              </div>
            </div>
          </ListItem>
          <Divider />

          <ListItem button dense onClick={ () => openDialog(Dialog.EditDueDateDialog) }>
            <ListItemIcon>
              <Event />
            </ListItemIcon>
            <div className="grid grid-cols-3 grid-rows-2 grid-flow-col gap-1 w-full">
              <div className="col-span-3">
                <Typography>
                  Due Date
                </Typography>
              </div>
              <div className="col-span-3 opacity-50">
                <Typography variant="body2">
                  { expense.description }
                </Typography>
              </div>
              <div className="col-span-1 row-span-2 flex justify-end">
                <ChevronRight className="align-middle h-full" />
              </div>
            </div>
          </ListItem>
          <Divider />

          <ListItem button dense onClick={ () => openDialog(Dialog.EditFundingScheduleDialog) }>
            <ListItemIcon>
              <ArrowForward />
            </ListItemIcon>
            <div className="grid grid-cols-3 grid-rows-2 grid-flow-col gap-1 w-full">
              <div className="col-span-3">
                <Typography>
                  Money In
                </Typography>
              </div>
              <div className="col-span-3 opacity-50">
                <Typography variant="body2">
                  { expense.getNextContributionAmountString() }/{ fundingSchedule?.name }
                </Typography>
              </div>
              <div className="col-span-1 row-span-2 flex justify-end">
                <ChevronRight className="align-middle h-full" />
              </div>
            </div>
          </ListItem>
          <Divider />

          <ListItem dense className="opacity-50">
            <ListItemIcon>
              <TrackChanges />
            </ListItemIcon>
            <div className="grid grid-cols-3 grid-rows-2 grid-flow-col gap-1 w-full">
              <div className="col-span-3">
                <Typography>
                  Contribution Option (WIP)
                </Typography>
              </div>
              <div className="col-span-3 opacity-50">
                <Typography variant="body2">
                  Set aside target amount
                </Typography>
              </div>
              <div className="col-span-1 row-span-2 flex justify-end">
                <ChevronRight className="align-middle h-full" />
              </div>
            </div>
          </ListItem>
          <Divider />

          <ListItem dense className="opacity-50">
            <ListItemIcon>
              <ArrowBack />
            </ListItemIcon>
            <div className="grid grid-cols-3 grid-rows-2 grid-flow-col gap-1 w-full">
              <div className="col-span-3">
                <Typography>
                  Money Out (WIP)
                </Typography>
              </div>
              <div className="col-span-3 opacity-50">
                <Typography variant="body2">
                  ....
                </Typography>
              </div>
              <div className="col-span-1 row-span-2 flex justify-end">
                <ChevronRight className="align-middle h-full" />
              </div>
            </div>
          </ListItem>
        </List>
        <div className="grid grid-cols-2 grid-flow-col mb-5">
          <div className="col-span-1">
            <Button variant="outlined" color="secondary" className="w-10/12" onClick={ deleteExpense }>
              <DeleteOutline className="mr-2" />
              Delete
            </Button>
          </div>
          <div className="col-span-1 flex justify-end">
            <Button variant="outlined" onClick={ () => openDialog(Dialog.TransferDialog) } className="w-10/12">
              <SwapHoriz className="mr-2" />
              Transfer
            </Button>
          </div>
        </div>
      </div>
    </Fragment>
  );
}
