import { Button, Card, Divider, IconButton, LinearProgress, List, ListItem, Typography } from '@material-ui/core';
import { ArrowBack, DeleteOutline } from '@material-ui/icons';
import NewGoalDialog from 'components/Goals/NewGoalDialog';
import TransferDialog from 'components/Spending/TransferDialog';
import FundingSchedule from 'data/FundingSchedule';
import Spending from 'data/Spending';
import moment from 'moment';
import React, { Component, Fragment } from 'react';
import { connect } from 'react-redux';
import { getFundingScheduleById } from 'shared/fundingSchedules/selectors/getFundingScheduleById';
import selectGoal from 'shared/spending/actions/selectGoal';
import { getSelectedGoal } from 'shared/spending/selectors/getSelectedGoal';

interface WithConnectionPropTypes {
  selectGoal: { (goalId: number | null): void }
  goal: Spending | null;
  fundingSchedule: FundingSchedule | null;
}

interface State {
  newGoalDialogOpen: boolean;
  transferDialogOpen: boolean;
}

export class GoalDetails extends Component<WithConnectionPropTypes, State> {

  state = {
    newGoalDialogOpen: false,
    transferDialogOpen: false,
  };

  openNewGoalDialog = () => this.setState({
    newGoalDialogOpen: true,
  });

  closeNewGoalDialog = () => this.setState({
    newGoalDialogOpen: false,
  });

  openTransferDialog = () => this.setState({
    transferDialogOpen: true,
  });

  closeTransferDialog = () => this.setState({
    transferDialogOpen: false,
  });

  renderInProgress = () => {
    const { goal, fundingSchedule } = this.props;

    const due = goal.nextRecurrence;

    // If the goal is the same year then just do the month and the day, but if its a different year then do the month
    // the day, and the year.
    const dueDate = due.year() !== moment().year() ? due.format('MMMM Do, YYYY') : due.format('MMMM Do')

    return (
      <div className="w-full h-full">
        <div className="w-full h-12">
          <div className="grid grid-cols-6 grid-rows-1 grid-flow-col">
            <div className="col-span-1">
              <IconButton
                onClick={ () => this.props.selectGoal(null) }
              >
                <ArrowBack/>
              </IconButton>
            </div>
            <div className="col-span-4 flex justify-center items-center">
              <Typography
                variant="h6"
              >
                In-progress Goal
              </Typography>
            </div>
            <div className="col-span-1">
              <IconButton disabled>
                <DeleteOutline/>
              </IconButton>
            </div>
          </div>
        </div>
        <Divider/>

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
        <Divider/>

        <div className="w-full pt-5 pb-5">
          <div className="grid grid-cols-3 grid-rows-3">
            <div className="h-5 col-span-2 row-span-1 flex justify-start">
              <Typography
                className="opacity-50"
                variant="caption"
              >
                Date Created (WIP)
              </Typography>
            </div>
            <div className="col-span-1 row-span-1 flex justify-end">
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
            <div className="col-span-1 row-span-1 flex flex-col justify-start">
              <Typography
                className="flex-1 flex justify-start"
                variant="caption"
              >
                <b>{ goal.getCurrentAmountString() }</b>
              </Typography>
              <Typography
                className="relative flex-1 flex top-1 justify-start"
                variant="caption"
              >
                Saved
              </Typography>
            </div>
            <div className="col-span-1 row-span-1 flex flex-col justify-center">
              <Typography
                className="flex-1 flex justify-center"
                variant="caption"
              >
                <b>{ goal.getNextContributionAmountString() }</b>
              </Typography>
              <Typography
                className="relative flex-1 flex top-1 justify-center"
                variant="caption"
              >
                on { fundingSchedule.name }
              </Typography>
            </div>
            <div className="col-span-1 row-span-1 flex flex-col justify-end">
              <Typography
                className="flex-1 flex justify-end"
                variant="caption"
              >
                <b>{ goal.getTargetAmountString() }</b>
              </Typography>
              <Typography
                className="relative flex-1 flex top-1 justify-end"
                variant="caption"
              >
                Target
              </Typography>
            </div>
          </div>
        </div>
        <Divider/>

        <div className="w-full pt-5 pb-5">
          <Button
            disabled
            color="secondary"
          >
            Pause Goal (WIP)
          </Button>
        </div>
        <Divider/>

        <div className="w-full pt-5 pb-5">
          <div className="grid grid-cols-3 grid-rows-2 grid-flow-col gap-1 opacity-50">
            <div className="col-span-2 row-span-1">
              <Typography
                variant="subtitle1"
              >
                Auto-spend (WIP)
              </Typography>
            </div>
            <div className="col-span-2 row-span-1 flex items-end">
              <Typography
                variant="subtitle2"
              >
                No categories selected
              </Typography>
            </div>
            <div className="col-span-1 row-span-2 flex justify-end items-center">
              <Button
                disabled
                color="primary"
              >
                Add
              </Button>
            </div>
          </div>
        </div>
        <Divider/>

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
                { fundingSchedule.name } · Next on { fundingSchedule.nextOccurrence.format('MMMM Do') }
              </Typography>
            </div>
          </div>
        </div>
        <Divider/>

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
              <Divider/>
              <ListItem key="wip" className="flex justify-center items-center opacity-50">
                <Typography>
                  Transactions For Thing (WIP)
                </Typography>
              </ListItem>
            </List>
          </Card>
        </div>
        <Divider/>

        <div className="w-full pt-5 pb-5 grid grid-cols-2 grid-flow-col gap-1">
          <div className="col-span-1 flex justify-start items-center">
            <Button
              variant="outlined"
            >
              More Edits
            </Button>
          </div>
          <div className="col-span-1 flex justify-end items-center">
            <Button
              variant="outlined"
              onClick={ this.openTransferDialog }
            >
              Transfer
            </Button>
          </div>
        </div>
      </div>
    );
  };

  renderComplete = () => {
    return null;
  };

  renderContents = () => {
    const { goal } = this.props;

    // If there is no goal selected then render the empty state view.
    if (!goal) {
      return this.renderNoGoal();
    }

    if (goal.getGoalIsInProgress()) {
      return this.renderInProgress();
    }

    return this.renderComplete();
  };

  renderNoGoal = () => {
    const { newGoalDialogOpen } = this.state;

    return (
      <Fragment>

        <div className="h-full flex justify-center items-center">
          <div className="grid grid-cols-1 grid-rows-2 grid-flow-col gap-2">
            <Typography
              className="opacity-50"
              variant="h6"
            >
              Select a goal, or create a new one...
            </Typography>
            <Button
              onClick={ this.openNewGoalDialog }
              color="primary"
            >
              Create A Goal
            </Button>
          </div>
        </div>
      </Fragment>
    )
  };

  renderDialogs = () => {
    const { goal } = this.props;
    const { newGoalDialogOpen, transferDialogOpen } = this.state;
    if (newGoalDialogOpen) {
      return <NewGoalDialog onClose={ this.closeNewGoalDialog } isOpen={ newGoalDialogOpen }/>;
    }

    if (transferDialogOpen) {
      return <TransferDialog isOpen onClose={ this.closeTransferDialog } initialToSpendingId={ goal.spendingId }/>;
    }

    return null;
  };

  render() {
    return (
      <div className="w-full h-full p-5">
        { this.renderDialogs() }
        { this.renderContents() }
      </div>
    );
  }
}

export default connect(
  (state) => {
    const goal = getSelectedGoal(state);
    return {
      goal,
      fundingSchedule: !!goal ? getFundingScheduleById(goal.fundingScheduleId)(state) : null,
    };
  },
  {
    selectGoal,
  }
)(GoalDetails);
