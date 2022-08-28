import React, { Fragment, useState } from 'react';
import { ArrowBack, DeleteOutline } from '@mui/icons-material';
import { Button, Card, Divider, IconButton, List, ListItem, Typography } from '@mui/material';

import TransferDialog from 'components/Spending/TransferDialog';
import { useFundingSchedule } from 'hooks/fundingSchedules';
import { useRemoveSpending } from 'hooks/spending';
import Spending from 'models/Spending';

interface Props {
  goal: Spending;
  onBack: () => void;
}

export default function CompletedGoalDetails(props: Props): JSX.Element {
  const { goal, onBack } = props;
  const fundingSchedule = useFundingSchedule(goal.fundingScheduleId);
  const removeSpending = useRemoveSpending();
  const [transferDialogOpen, setTransferDialogOpen] = useState(false);

  function TransferDialogMaybe(): JSX.Element {
    if (!transferDialogOpen) {
      return null;
    }

    function closeTransferDialog() {
      setTransferDialogOpen(false);
    }

    return <TransferDialog isOpen onClose={ closeTransferDialog } initialToSpendingId={ goal.spendingId } />;
  }

  function openTransferDialog() {
    setTransferDialogOpen(true);
  }

  async function deleteGoal(): Promise<void> {
    if (!goal) {
      return Promise.resolve();
    }

    if (window.confirm(`Are you sure you want to delete goal: ${ goal.name }`)) {
      return removeSpending(goal.spendingId)
        .then(() => void onBack());
    }

    return Promise.resolve();
  }

  return (
    <Fragment>
      <TransferDialogMaybe />
      <div className="w-full h-full">
        <div className="w-full h-12">
          <div className="grid grid-cols-6 grid-rows-1 grid-flow-col">
            <div className="col-span-1">
              <IconButton
                onClick={ onBack }
              >
                <ArrowBack />
              </IconButton>
            </div>
            <div className="col-span-4 flex justify-center items-center">
              <Typography
                variant="h6"
              >
                Completed Goal
              </Typography>
            </div>
            <div className="col-span-1">
              <IconButton onClick={ deleteGoal }>
                <DeleteOutline />
              </IconButton>
            </div>
          </div>
        </div>
        <Divider />

        <div className="w-full pt-5">
          <div className="w-full">
            <Card elevation={ 3 } className="h-32 w-full flex justify-center items-center">
              <Typography
                className="opacity-50"
              >
                Image here or something (WIP)
              </Typography>
            </Card>
          </div>
          <div className="w-full pt-2.5">
            <Typography
              variant="h6"
            >
              { goal.name }
            </Typography>
          </div>
        </div>
        <Divider />

        <div className="w-full pt-5 pb-5">
          <div className="grid grid-cols-1 grid-rows-2 grid-flow-col gap-1">
            <div className="col-span-1 row-span-1">
              <Typography
                variant="subtitle1"
              >
                Funding Schedule
              </Typography>
            </div>
            <div className="col-span-1 row-span-1 flex items-end">
              <Typography
                variant="subtitle2"
              >
                { fundingSchedule.name }
              </Typography>
            </div>
          </div>
        </div>
        <Divider />

        <div className="w-full pt-5 pb-5">
          <Card elevation={ 3 }>
            <List dense>
              <ListItem key="totals" className="grid grid-cols-3 grid-flow-col">
                <div className="col-span-2 flex justify-start items-center">
                  <Typography>
                    Total spent from Goal
                  </Typography>
                </div>
                <div className="col-span-1 flex justify-end items-center">
                  <Typography>
                    { goal.getUsedAmountString() }
                  </Typography>
                </div>
              </ListItem>
              <Divider />
              <ListItem key="wip" className="flex justify-center items-center opacity-50">
                <Typography>
                  Transactions For Thing (WIP)
                </Typography>
              </ListItem>
            </List>
          </Card>
        </div>
        <Divider />

        <div className="w-full pt-5 pb-5 grid grid-cols-2 grid-flow-col gap-1">
          <div className="col-span-2 flex justify-end items-center">
            <Button
              variant="outlined"
              onClick={ openTransferDialog }
            >
              Transfer
            </Button>
          </div>
        </div>
      </div>
    </Fragment>
  );
}
