import React, { Fragment } from 'react';
import { Checkbox, Chip, LinearProgress, ListItem, ListItemIcon, Typography } from '@mui/material';
import moment from 'moment';
import shallow from 'zustand/shallow';

import { useFundingSchedule } from 'hooks/fundingSchedules';
import useStore from 'hooks/store';
import Spending from 'models/Spending';

import './styles/GoalRow.scss';

interface Props {
  goal: Spending;
}

export default function GoalRow(props: Props): JSX.Element  {
  const { selectedGoalId, setCurrentGoal } = useStore(state => ({
    selectedGoalId: state.selectedGoalId,
    setCurrentGoal: state.setCurrentGoal,
  }), shallow);
  const isSelected = props.goal.spendingId === selectedGoalId;
  const fundingSchedule = useFundingSchedule(props.goal.fundingScheduleId);

  function InProgress(): JSX.Element {
    const { goal } = props;
    const due = goal.nextRecurrence;

    // If the goal is the same year then just do the month and the day, but if its a different year then do the month
    // the day, and the year.
    const date = due.year() !== moment().year() ? due.format('MMMM Do, YYYY') : due.format('MMMM Do');

    return (
      <div className="w-full grid grid-cols-3 grid-rows-3 grid-flow-col">
        <div className="col-span-3">
          <Typography
            variant="subtitle1"
          >
            { props.goal.name }
          </Typography>
        </div>
        <div className="flex items-center col-span-3">
          <LinearProgress
            classes={ {
              buffer: 'MuiLinearProgress-colorPrimary',
            } }
            className="w-full"
            variant="buffer"
            color="primary"
            valueBuffer={ ((goal.currentAmount + goal.usedAmount) / goal.targetAmount) * 100 }
            value={ (goal.usedAmount / goal.targetAmount) * 100 }
          />
        </div>
        <div className="col-span-1">
          <Typography
            variant="body2"
          >
            <b>{ goal.getCurrentAmountString() }</b>
          </Typography>
        </div>
        <div className="flex justify-center col-span-1">
          <Typography
            variant="body2"
          >
            { goal.isPaused && 'Paused' }
            { !goal.isPaused &&
              <Fragment>
                <b>{ goal.getNextContributionAmountString() }</b> on { fundingSchedule?.name }
              </Fragment>
            }
          </Typography>
        </div>
        <div className="flex justify-end col-span-1">
          <Typography
            variant="body2"
          >
            <b>{ goal.getTargetAmountString() }</b> by { date }
          </Typography>
        </div>
      </div>
    );
  }

  function Complete(): JSX.Element {
    const { goal } = props;

    return (
      <div className="w-full h-full grid grid-cols-4 grid-rows-1 grid-flow-col gap-1">
        <div className="col-span-3">
          <Typography
            variant="subtitle1"
          >
            { goal.name }
          </Typography>
        </div>
        <div className="flex justify-end col-span-1">
          <Chip
            label={ goal.getCurrentAmountString() }
            color="primary"
          />
        </div>
      </div>
    );
  }

  function Contents(): JSX.Element {
    const { goal } = props;

    if (goal.getGoalIsInProgress()) {
      return <InProgress />;
    }

    return <Complete />;
  }

  function onClick() {
    setCurrentGoal(selectedGoalId === goal.spendingId ? null : goal.spendingId);
  }

  const { goal } = props;
  if (!goal) {
    return null;
  }

  return (
    <ListItem
      button
      className="goal-row"
      onClick={ onClick }
    >
      <ListItemIcon>
        <Checkbox
          edge="start"
          checked={ isSelected }
          tabIndex={ -1 }
          color="primary"
        />
      </ListItemIcon>
      <Contents />
    </ListItem>
  );
}
