import React, { Fragment } from 'react';
import { ArrowBack, DeleteOutline } from '@mui/icons-material';
import { Button, Card, Divider, IconButton, LinearProgress, List, ListItem, Typography } from '@mui/material';
import moment from 'moment';

import { useFundingSchedule } from 'hooks/fundingSchedules';
import { useRemoveSpending, useUpdateSpending } from 'hooks/spending';
import Spending from 'models/Spending';

interface Props {
  goal: Spending;
  onBack: () => void;
  openEditView: () => void;
  openTransferDialog: () => void;
}

export default function InProgressGoalDetails(props: Props): JSX.Element {
  const { goal, onBack, openEditView, openTransferDialog } = props;
  const fundingSchedule = useFundingSchedule(goal.fundingScheduleId);
  const removeSpending = useRemoveSpending();
  const updateSpending = useUpdateSpending();

  function deleteGoal(): Promise<void> {
    if (!goal) {
      return Promise.resolve();
    }

    if (window.confirm(`Are you sure you want to delete goal: ${ goal.name }`)) {
      return removeSpending(goal.spendingId);
    }

    return Promise.resolve();
  }


  function togglePauseGoal(): Promise<void> {
    if (!goal) {
      return Promise.resolve();
    }

    const updatedGoal = new Spending({
      ...goal,
      isPaused: !goal.isPaused,
    });

    return updateSpending(updatedGoal);
  }

  const created = goal.dateCreated;
  const due = goal.nextRecurrence;

  // If the goal is the same year then just do the month and the day, but if its a different year then do the month
  // the day, and the year.
  const dueDate = due.year() !== moment().year() ? due.format('MMMM Do, YYYY') : due.format('MMMM Do');
  const createdDate = created.year() !== moment().year() ? created.format('MMMM Do, YYYY') : created.format('MMMM Do');

  return (
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
          <div className="flex items-center justify-center col-span-4">
            <Typography
              variant="h6"
            >
              In-progress Goal
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
          <Card elevation={ 3 } className="flex items-center justify-center w-full h-32">
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
        <div className="grid grid-cols-3 grid-rows-3">
          <div className="flex justify-start h-5 col-span-2 row-span-1">
            <Typography
              variant="caption"
            >
              { createdDate }
            </Typography>
          </div>
          <div className="flex justify-end col-span-1 row-span-1">
            <Typography
              variant="caption"
            >
              { dueDate }
            </Typography>
          </div>
          <div className="col-span-3 row-span-1">
            <LinearProgress
              classes={ {
                buffer: 'MuiLinearProgress-colorPrimary',
              } }
              className="w-full goal-progress"
              variant="buffer"
              color="primary"
              valueBuffer={ ((goal.currentAmount + goal.usedAmount) / goal.targetAmount) * 100 }
              value={ (goal.usedAmount / goal.targetAmount) * 100 }
            />
          </div>
          <div className="flex flex-col justify-start col-span-1 row-span-1">
            <Typography
              className="flex justify-start flex-1"
              variant="caption"
            >
              <b>{ goal.getCurrentAmountString() }</b>
            </Typography>
            <Typography
              className="relative flex justify-start flex-1 top-1"
              variant="caption"
            >
              Saved
            </Typography>
          </div>
          <div className="flex flex-col justify-center col-span-1 row-span-1">
            { !goal.isPaused &&
              <Fragment>
                <Typography
                  className="flex justify-center flex-1"
                  variant="caption"
                >
                  <b>{ goal.getNextContributionAmountString() }</b>
                </Typography>
                <Typography
                  className="relative flex justify-center flex-1 top-1"
                  variant="caption"
                >
                  on { fundingSchedule.name }
                </Typography>
              </Fragment>
            }
            { goal.isPaused &&
              <Fragment>
                <Typography
                  className="flex justify-center flex-1"
                  variant="body2"
                >
                  Paused
                </Typography>
              </Fragment> }
          </div>
          <div className="flex flex-col justify-end col-span-1 row-span-1">
            <Typography
              className="flex justify-end flex-1"
              variant="caption"
            >
              <b>{ goal.getTargetAmountString() }</b>
            </Typography>
            <Typography
              className="relative flex justify-end flex-1 top-1"
              variant="caption"
            >
              Target
            </Typography>
          </div>
        </div>
      </div>
      <Divider />

      <div className="w-full pt-5 pb-5">
        <Button
          onClick={ togglePauseGoal }
          color="secondary"
        >
          { goal.isPaused ? 'Unpause Goal ' : 'Pause Goal' }
        </Button>
      </div>
      <Divider />

      <div className="w-full pt-5 pb-5">
        <div className="opacity-50 grid grid-cols-3 grid-rows-2 grid-flow-col gap-1">
          <div className="col-span-2 row-span-1">
            <Typography
              variant="subtitle1"
            >
              Auto-spend (WIP)
            </Typography>
          </div>
          <div className="flex items-end col-span-2 row-span-1">
            <Typography
              variant="subtitle2"
            >
              No categories selected
            </Typography>
          </div>
          <div className="flex items-center justify-end col-span-1 row-span-2">
            <Button
              disabled
              color="primary"
            >
              Add
            </Button>
          </div>
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
          <div className="flex items-end col-span-1 row-span-1">
            <Typography
              variant="subtitle2"
            >
              { fundingSchedule.name } · Next on { fundingSchedule.nextOccurrence.format('MMMM Do') }
            </Typography>
          </div>
        </div>
      </div>
      <Divider />

      <div className="w-full pt-5 pb-5">
        <Card elevation={ 3 }>
          <List dense>
            <ListItem key="totals" className="grid grid-cols-3 grid-flow-col">
              <div className="flex items-center justify-start col-span-2">
                <Typography>
                  Total spent from Goal
                </Typography>
              </div>
              <div className="flex items-center justify-end col-span-1">
                <Typography>
                  { goal.getUsedAmountString() }
                </Typography>
              </div>
            </ListItem>
            <Divider />
            <ListItem key="wip" className="flex items-center justify-center opacity-50">
              <Typography>
                Transactions For Thing (WIP)
              </Typography>
            </ListItem>
          </List>
        </Card>
      </div>
      <Divider />

      <div className="w-full pt-5 pb-5 grid grid-cols-2 grid-flow-col gap-1">
        <div className="flex items-center justify-start col-span-1">
          <Button
            variant="outlined"
            onClick={ openEditView }
          >
            More Edits
          </Button>
        </div>
        <div className="flex items-center justify-end col-span-1">
          <Button
            variant="outlined"
            onClick={ openTransferDialog }
          >
            Transfer
          </Button>
        </div>
      </div>
    </div>
  );
}
