import React, { Fragment } from 'react';
import { Button, Typography } from '@mui/material';

import { showCreateGoalDialog } from 'components/Goals/CreateGoalDialog';

export default function NoGoals(): JSX.Element {
  return (
    <Fragment>
      <div className="flex items-center justify-center h-full">
        <div className="grid grid-cols-1 grid-rows-2 grid-flow-col gap-2">
          <Typography
            className="opacity-50"
            variant="h6"
          >
              Select a goal, or create a new one...
          </Typography>
          <Button
            onClick={ showCreateGoalDialog }
            color="primary"
          >
              Create A Goal
          </Button>
        </div>
      </div>
    </Fragment>
  );
}
