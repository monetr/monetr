import React, { Fragment, useState } from 'react';
import { Button, Typography } from '@mui/material';

import NewGoalDialog from 'components/Goals/NewGoalDialog';

export default function NoGoals(): JSX.Element {
  const [newGoalDialogOpen, setNewGoalDialogOpen] = useState(false);

  function openNewGoalDialog() {
    setNewGoalDialogOpen(true);
  }

  function NewGoalDialogMaybe(): JSX.Element {
    if (!newGoalDialogOpen) {
      return null;
    }

    function closeNewGoalDialog() {
      setNewGoalDialogOpen(false);
    }

    return <NewGoalDialog onClose={ closeNewGoalDialog } isOpen />;
  }
  
  return (
    <Fragment>
      <NewGoalDialogMaybe />
      <div className="flex items-center justify-center h-full">
        <div className="grid grid-cols-1 grid-rows-2 grid-flow-col gap-2">
          <Typography
            className="opacity-50"
            variant="h6"
          >
              Select a goal, or create a new one...
          </Typography>
          <Button
            onClick={ openNewGoalDialog }
            color="primary"
          >
              Create A Goal
          </Button>
        </div>
      </div>
    </Fragment>
  );
}
