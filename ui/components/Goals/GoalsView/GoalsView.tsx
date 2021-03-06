import { Button, Card, Divider, List, Typography } from '@mui/material';
import GoalDetails from 'components/Goals/GoalDetails';
import GoalRow from 'components/Goals/GoalRow';
import NewGoalDialog from 'components/Goals/NewGoalDialog';
import React, { Fragment, useState } from 'react';
import { useSelector } from 'react-redux';
import { getGoalIds } from 'shared/spending/selectors/getGoalIds';

import 'components/Goals/GoalsView/styles/GoalsView.scss';

export default function GoalsView(): JSX.Element {
  const [newGoalDialogOpen, setNewGoalDialogOpen] = useState(false);
  const goalIds = useSelector(getGoalIds);

  function openNewGoalDialog() {
    setNewGoalDialogOpen(true);
  }

  function closeNewGoalDialog() {
    setNewGoalDialogOpen(false);
  }

  function GoalList(): JSX.Element {
    return (
      <div className="w-full goals-list">
        <List disablePadding className="w-full">
          {
            goalIds.map(item => (
              <Fragment key={ item }>
                <GoalRow goalId={ item }/>
                <Divider/>
              </Fragment>
            ))
          }
        </List>
      </div>
    );
  }

  if (goalIds.length === 0) {
    return (
      <Fragment>
        { newGoalDialogOpen && <NewGoalDialog onClose={ closeNewGoalDialog } isOpen={ newGoalDialogOpen }/> }
        <div className="minus-nav bg-primary">
          <div className="flex flex-col h-full max-h-full bg-white view-inner">
            <div className="grid grid-cols-3 flex-grow">
              <div className="col-span-3">
                <div className="w-full goals-list ">
                  <div className="h-full flex justify-center items-center">
                    <div className="grid grid-cols-1 grid-rows-2 grid-flow-col gap-2">
                      <Typography
                        className="opacity-50"
                        variant="h3"
                      >
                        You don't have any goals yet...
                      </Typography>
                      <Button
                        onClick={ openNewGoalDialog }
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
              </div>
            </div>
          </div>
        </div>
      </Fragment>
    )
  }

  return (
    <div className="minus-nav bg-primary">
      <div className="flex flex-col h-full max-h-full view-inner">
        <div className="grid grid-cols-3 flex-grow">
          <div className="col-span-2">
            <GoalList/>
          </div>
          <div className="border-l w-full goals-list">
            <GoalDetails/>
          </div>
        </div>
      </div>
    </div>
  );
}
