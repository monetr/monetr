import React, { Fragment } from 'react';
import { Button, Divider, List, Typography } from '@mui/material';

import { showCreateGoalDialog } from 'components/Goals/CreateGoalDialog';
import GoalDetails from 'components/Goals/GoalDetails';
import GoalRow from 'components/Goals/GoalRow';
import { useSpendingFiltered } from 'hooks/spending';
import { SpendingType } from 'models/Spending';

import 'components/Goals/styles/GoalsView.scss';

export default function GoalsView(): JSX.Element {
  const { result: goals } = useSpendingFiltered(SpendingType.Goal);

  function GoalList(): JSX.Element {
    return (
      <div className="w-full goals-list">
        <List disablePadding className="w-full">
          {
            goals
              .sort((a, b) => a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1)
              .map(item => (
                <Fragment key={ item.spendingId }>
                  <GoalRow goal={ item } />
                  <Divider />
                </Fragment>
              ))
          }
        </List>
      </div>
    );
  }

  if (goals.length === 0) {
    return (
      <div className="h-full w-full bg-primary">
        <div className="view-inner h-full flex justify-center items-center">
          <div className="grid grid-cols-1 grid-rows-2 grid-flow-col gap-2">
            <Typography
              className="opacity-50"
              variant="h3"
            >
              You don't have any goals yet...
            </Typography>
            <Button
              onClick={ showCreateGoalDialog }
              color="primary"
            >
              <Typography
                variant="h6"
              >
                Create A Goal
              </Typography>
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="minus-nav bg-primary">
      <div className="flex flex-col h-full max-h-full view-inner">
        <div className="grid grid-cols-3 flex-grow">
          <div className="col-span-2">
            <GoalList />
          </div>
          <div className="border-l w-full goals-list">
            <GoalDetails />
          </div>
        </div>
      </div>
    </div>
  );
}
