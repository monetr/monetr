import { Checkbox, Chip, LinearProgress, ListItem, ListItemIcon, Typography } from '@material-ui/core';
import FundingSchedule from 'data/FundingSchedule';
import Spending from 'data/Spending';
import moment from 'moment';
import React, { Fragment, Component } from 'react';
import { connect } from 'react-redux';
import { getFundingScheduleById } from 'shared/fundingSchedules/selectors/getFundingScheduleById';
import selectGoal from 'shared/spending/actions/selectGoal';
import { getGoalIsSelected } from 'shared/spending/selectors/getGoalIsSelected';
import { getSpendingById } from 'shared/spending/selectors/getSpendingById';

import './styles/GoalRow.scss';

export interface PropTypes {
  goalId: number;
}

interface WithConnectionPropTypes extends PropTypes {
  isSelected: boolean;
  goal: Spending;
  fundingSchedule: FundingSchedule;
  selectGoal: { (goalId: number): void }
}

export class GoalRow extends Component<WithConnectionPropTypes, any> {

  onClick = () => {
    return this.props.selectGoal(this.props.goalId);
  };

  renderInProgress = () => {
    const { goal, fundingSchedule } = this.props;

    const due = goal.nextRecurrence;

    // If the goal is the same year then just do the month and the day, but if its a different year then do the month
    // the day, and the year.
    const date = due.year() !== moment().year() ? due.format('MMMM Do, YYYY') : due.format('MMMM Do')

    return (
      <div className="grid grid-cols-3 grid-rows-3 grid-flow-col w-full">
        <div className="col-span-3">
          <Typography
            variant="subtitle1"
          >
            { goal.name }
          </Typography>
        </div>
        <div className="col-span-3 flex items-center">
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
        <div className="col-span-1 flex justify-center">
          <Typography
            variant="body2"
          >
            { goal.isPaused && 'Paused' }
            { !goal.isPaused &&
            <Fragment>
              <b>{ goal.getNextContributionAmountString() }</b> on { fundingSchedule.name }
            </Fragment>
            }
          </Typography>
        </div>
        <div className="col-span-1 flex justify-end">
          <Typography
            variant="body2"
          >
            <b>{ goal.getTargetAmountString() }</b> by { date }
          </Typography>
        </div>
      </div>
    )
  };

  renderComplete = () => {
    const { goal } = this.props;

    return (
      <div className="grid grid-cols-4 grid-rows-1 grid-flow-col gap-1 w-full h-full">
        <div className="col-span-3">
          <Typography
            variant="subtitle1"
          >
            { goal.name }
          </Typography>
        </div>
        <div className="col-span-1 flex justify-end">
          <Chip
            label={ goal.getCurrentAmountString() }
            color="primary"
          />
        </div>
      </div>
    )
  };

  renderContents = () => {
    const { goal } = this.props;

    if (goal.getGoalIsInProgress()) {
      return this.renderInProgress();
    }

    return this.renderComplete();
  };

  render() {
    const { isSelected } = this.props;

    return (
      <ListItem
        button
        className="goal-row"
        onClick={ this.onClick }
      >
        <ListItemIcon>
          <Checkbox
            edge="start"
            checked={ isSelected }
            tabIndex={ -1 }
            color="primary"
          />
        </ListItemIcon>
        { this.renderContents() }
      </ListItem>
    );
  }
}

export default connect(
  (state, props: PropTypes) => {
    const goal = getSpendingById(props.goalId)(state);
    return {
      goal,
      fundingSchedule: getFundingScheduleById(goal.fundingScheduleId)(state),
      isSelected: getGoalIsSelected(props.goalId)(state),
    }
  },
  {
    selectGoal,
  }
)(GoalRow);
