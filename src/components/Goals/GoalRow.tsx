import { Checkbox, Chip, LinearProgress, ListItem, ListItemIcon, Typography } from '@material-ui/core';
import FundingSchedule from 'data/FundingSchedule';
import Spending from 'data/Spending';
import React, { Component } from 'react';
import { connect } from 'react-redux';
import { getFundingScheduleById } from 'shared/fundingSchedules/selectors/getFundingScheduleById';
import selectGoal from 'shared/spending/actions/selectGoal';
import { getGoalIsSelected } from 'shared/spending/selectors/getGoalIsSelected';
import { getSpendingById } from 'shared/spending/selectors/getSpendingById';

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

    return (
      <div className="grid grid-cols-3 grid-rows-3 grid-flow-col gap-1 w-full">
        <div className="col-span-3">
          <Typography
            variant="subtitle1"
          >
            { goal.name }
          </Typography>
        </div>
        <div className="col-span-3">
          <LinearProgress
            variant="determinate"
            color="primary"
            value={ ((goal.currentAmount + goal.usedAmount) / goal.targetAmount) * 100 }
            // TODO valueBuffer might only work if the variant is buffer.
            //  If this is the case then we will want to have a CSS rule to suppress the dotty bois.
            valueBuffer={ (goal.usedAmount / goal.targetAmount) * 100 }
          />
        </div>
        <div className="col-span-1">
          <Typography
            variant="body2"
          >
            <b>{ goal.getGoalSavedAmountString() }</b>
          </Typography>
        </div>
        <div className="col-span-1 flex justify-center">
          <Typography
            variant="body2"
          >
            <b>{ goal.getNextContributionAmountString() }</b> on { fundingSchedule.name }
          </Typography>
        </div>
        <div className="col-span-1 flex justify-end">
          <Typography
            variant="body2"
          >
            <b>{ goal.getTargetAmountString() }</b> by { goal.nextRecurrence.format('MMMM Do') }
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
