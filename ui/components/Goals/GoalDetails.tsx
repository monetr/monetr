import { Button, Card, Divider, IconButton, LinearProgress, List, ListItem, Typography } from '@mui/material';
import { ArrowBack, DeleteOutline } from '@mui/icons-material';
import NewGoalDialog from 'components/Goals/NewGoalDialog';
import TransferDialog from 'components/Spending/TransferDialog';
import FundingSchedule from 'models/FundingSchedule';
import Spending from 'models/Spending';
import moment from 'moment';
import React, { Component, Fragment } from 'react';
import { connect } from 'react-redux';
import { getFundingScheduleById } from 'shared/fundingSchedules/selectors/getFundingScheduleById';
import selectGoal from 'shared/spending/actions/selectGoal';
import { getSelectedGoal } from 'shared/spending/selectors/getSelectedGoal';
import EditGoalView from "components/Goals/EditGoalView";
import deleteSpending from "shared/spending/actions/deleteSpending";
import fetchBalances from "shared/balances/actions/fetchBalances";
import updateSpending from "shared/spending/actions/updateSpending";

interface WithConnectionPropTypes {
  selectGoal: { (goalId: number | null): void }
  deleteSpending: (spending: Spending) => Promise<void>;
  updateSpending: (spending: Spending) => Promise<void>;
  fetchBalances: () => Promise<void>;
  goal: Spending | null;
  fundingSchedule: FundingSchedule | null;
}

interface State {
  newGoalDialogOpen: boolean;
  transferDialogOpen: boolean;
  editGoalOpen: boolean;
}

export class GoalDetails extends Component<WithConnectionPropTypes, State> {

  state = {
    newGoalDialogOpen: false,
    transferDialogOpen: false,
    editGoalOpen: false,
  };

  componentDidUpdate(prevProps: Readonly<WithConnectionPropTypes>, prevState: Readonly<State>, snapshot?: any) {
    // When the user "un-selects" a goal, we want to make sure that we close the edit goal view.
    if (this.props.goal?.spendingId !== prevProps.goal?.spendingId && this.state.editGoalOpen) {
      this.setState({
        editGoalOpen: false,
      });
    }
  }

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

  openEditView = () => this.setState({
    editGoalOpen: true,
  });

  closeEditView = () => this.setState({
    editGoalOpen: false,
  });

  deleteGoal = () => {
    const { goal } = this.props;
    if (!goal) {
      return Promise.resolve();
    }

    if (window.confirm(`Are you sure you want to delete goal: ${ goal.name }`)) {
      return this.props.deleteSpending(goal).then(() => this.props.fetchBalances());
    }

    return Promise.resolve();
  };

  togglePauseGoal = () => {
    const { goal } = this.props;
    if (!goal) {
      return Promise.resolve();
    }

    const updatedGoal = new Spending({
      ...goal,
      isPaused: !goal.isPaused
    });

    return this.props.updateSpending(updatedGoal);
  };

  renderInProgress = () => {
    const { goal, fundingSchedule } = this.props;

    const created = goal.dateCreated;
    const due = goal.nextRecurrence;

    // If the goal is the same year then just do the month and the day, but if its a different year then do the month
    // the day, and the year.
    const dueDate = due.year() !== moment().year() ? due.format('MMMM Do, YYYY') : due.format('MMMM Do')
    const createdDate = created.year() !== moment().year() ? created.format('MMMM Do, YYYY') : created.format('MMMM Do');

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
            <div className="flex items-center justify-center col-span-4">
              <Typography
                variant="h6"
              >
                In-progress Goal
              </Typography>
            </div>
            <div className="col-span-1">
              <IconButton onClick={ this.deleteGoal }>
                <DeleteOutline/>
              </IconButton>
            </div>
          </div>
        </div>
        <Divider/>

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
        <Divider/>

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
        <Divider/>

        <div className="w-full pt-5 pb-5">
          <Button
            onClick={ this.togglePauseGoal }
            color="secondary"
          >
            { goal.isPaused ? 'Unpause Goal ' : 'Pause Goal' }
          </Button>
        </div>
        <Divider/>

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
            <div className="flex items-end col-span-1 row-span-1">
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
              <Divider/>
              <ListItem key="wip" className="flex items-center justify-center opacity-50">
                <Typography>
                  Transactions For Thing (WIP)
                </Typography>
              </ListItem>
            </List>
          </Card>
        </div>
        <Divider/>

        <div className="w-full pt-5 pb-5 grid grid-cols-2 grid-flow-col gap-1">
          <div className="flex items-center justify-start col-span-1">
            <Button
              variant="outlined"
              onClick={ this.openEditView }
            >
              More Edits
            </Button>
          </div>
          <div className="flex items-center justify-end col-span-1">
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

    const { editGoalOpen } = this.state;
    if (editGoalOpen) {
      return (
        <EditGoalView hideView={ this.closeEditView }/>
      );
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

        <div className="flex items-center justify-center h-full">
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
      fundingSchedule: goal && getFundingScheduleById(goal.fundingScheduleId)(state),
    };
  },
  {
    selectGoal,
    deleteSpending,
    fetchBalances,
    updateSpending,
  }
)(GoalDetails);
