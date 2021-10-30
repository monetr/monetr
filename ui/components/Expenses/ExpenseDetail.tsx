import { Button, Divider, LinearProgress, List, ListItem, ListItemIcon, Typography } from '@material-ui/core';
import {
  AccountBalance,
  ArrowBack,
  ArrowForward,
  ChevronRight,
  DeleteOutline,
  Event,
  SwapHoriz,
  TrackChanges
} from '@material-ui/icons';
import TransferDialog from 'components/Spending/TransferDialog';
import FundingSchedule from 'models/FundingSchedule';
import Spending from 'models/Spending';
import { Map } from 'immutable';
import React, { Component, Fragment } from 'react';
import { connect } from 'react-redux';
import { getFundingSchedules } from 'shared/fundingSchedules/selectors/getFundingSchedules';
import { getSelectedExpense } from 'shared/spending/selectors/getSelectedExpense';
import EditSpendingAmountDialog from "components/Expenses/EditExpenseAmountDialog";
import EditExpenseDueDateDialog from "components/Expenses/EditExpenseDueDateDialog";
import FundingScheduleList from "components/FundingSchedules/FundingScheduleList";
import fetchBalances from "shared/balances/actions/fetchBalances";
import deleteSpending from "shared/spending/actions/deleteSpending";
import EditExpenseFundingScheduleDialog from "components/Expenses/EditExpenseFundingScheduleDialog";

interface WithConnectionPropTypes {
  expense?: Spending;
  fundingSchedules: Map<number, FundingSchedule>;
  deleteSpending: (spending: Spending) => Promise<void>;
  fetchBalances: () => Promise<void>;
}

interface State {
  transferDialogOpen: boolean;
  editAmountDialogOpen: boolean;
  editDueDateDialogOpen: boolean;
  editFundingScheduleDialogOpen: boolean;
}

export class ExpenseDetail extends Component<WithConnectionPropTypes, State> {

  state = {
    transferDialogOpen: false,
    editAmountDialogOpen: false,
    editDueDateDialogOpen: false,
    editFundingScheduleDialogOpen: false,
  };

  openTransferDialog = () => {
    return this.setState({
      transferDialogOpen: true,
    });
  };

  closeTransferDialog = () => {
    return this.setState({
      transferDialogOpen: false,
    });
  };

  openEditAmountDialog = () => this.setState({
    editAmountDialogOpen: true,
  });

  closeEditAmountDialog = () => this.setState({
    editAmountDialogOpen: false,
  });

  openEditDueDateDialog = () => this.setState({
    editDueDateDialogOpen: true,
  });

  closeEditDueDateDialog = () => this.setState({
    editDueDateDialogOpen: false,
  });

  openEditFundingScheduleDialog = () => this.setState({
    editFundingScheduleDialogOpen: true,
  });

  closeEditFundingScheduleDialog = () => this.setState({
    editFundingScheduleDialogOpen: false,
  });

  renderNoExpenseSelected = () => {
    return (
      <FundingScheduleList/>
    )
  };

  deleteExpense = () => {
    const { expense } = this.props;
    if (!expense) {
      return Promise.resolve();
    }

    if (window.confirm(`Are you sure you want to delete expense: ${ expense.name }`)) {
      return this.props.deleteSpending(expense).then(() => this.props.fetchBalances());
    }

    return Promise.resolve();
  };

  render() {
    const { expense } = this.props;
    if (!expense) {
      return this.renderNoExpenseSelected();
    }

    const fundingSchedule = this.props.fundingSchedules.get(expense.fundingScheduleId, new FundingSchedule());

    const {
      transferDialogOpen,
      editAmountDialogOpen,
      editDueDateDialogOpen,
      editFundingScheduleDialogOpen,
    } = this.state;

    return (
      <Fragment>
        { transferDialogOpen &&
        <TransferDialog isOpen onClose={ this.closeTransferDialog } initialToSpendingId={ expense.spendingId }/>
        }
        { editAmountDialogOpen &&
        <EditSpendingAmountDialog isOpen onClose={ this.closeEditAmountDialog }/>
        }
        { editDueDateDialogOpen &&
        <EditExpenseDueDateDialog isOpen onClose={ this.closeEditDueDateDialog }/>
        }
        { editFundingScheduleDialogOpen &&
        <EditExpenseFundingScheduleDialog onClose={ this.closeEditFundingScheduleDialog } isOpen/>
        }

        <div className="w-full pl-5 pr-5 pt-5 expense-detail">
          <div className="grid grid-cols-3 grid-rows-4 grid-flow-col gap-1 w-auto">
            <div className="col-span-2">
              <Typography
                variant="h5"
              >
                { expense.name }
              </Typography>
            </div>
            <div className="col-span-2">
              <Typography
                variant="h6"
              >
                { expense.getCurrentAmountString() } of { expense.getTargetAmountString() }
              </Typography>
            </div>
            <div className="col-span-3">
              <Typography>
                { expense.getNextOccurrenceString() } - { expense.description }
              </Typography>
            </div>
            <div className="col-span-3">
              <Typography>
                { expense.getNextContributionAmountString() }/{ fundingSchedule.name }
              </Typography>
            </div>
            <div className="col-span-1 row-span-2">
              <LinearProgress
                className="mt-3"
                variant="determinate"
                value={ (expense.currentAmount / expense.targetAmount) * 100 }
              />
            </div>
          </div>

          <List dense>
            <Divider/>
            <ListItem button dense onClick={ this.openEditAmountDialog }>
              <ListItemIcon>
                <AccountBalance/>
              </ListItemIcon>
              <div className="grid grid-cols-3 grid-rows-2 grid-flow-col gap-1 w-full">
                <div className="col-span-3">
                  <Typography>
                    Amount
                  </Typography>
                </div>
                <div className="col-span-3 opacity-50">
                  <Typography variant="body2">
                    { expense.getTargetAmountString() }
                  </Typography>
                </div>
                <div className="col-span-1 row-span-2 flex justify-end">
                  <ChevronRight className="align-middle h-full"/>
                </div>
              </div>
            </ListItem>
            <Divider/>

            <ListItem button dense onClick={ this.openEditDueDateDialog }>
              <ListItemIcon>
                <Event/>
              </ListItemIcon>
              <div className="grid grid-cols-3 grid-rows-2 grid-flow-col gap-1 w-full">
                <div className="col-span-3">
                  <Typography>
                    Due Date
                  </Typography>
                </div>
                <div className="col-span-3 opacity-50">
                  <Typography variant="body2">
                    { expense.description }
                  </Typography>
                </div>
                <div className="col-span-1 row-span-2 flex justify-end">
                  <ChevronRight className="align-middle h-full"/>
                </div>
              </div>
            </ListItem>
            <Divider/>

            <ListItem button dense onClick={ this.openEditFundingScheduleDialog }>
              <ListItemIcon>
                <ArrowForward/>
              </ListItemIcon>
              <div className="grid grid-cols-3 grid-rows-2 grid-flow-col gap-1 w-full">
                <div className="col-span-3">
                  <Typography>
                    Money In
                  </Typography>
                </div>
                <div className="col-span-3 opacity-50">
                  <Typography variant="body2">
                    { expense.getNextContributionAmountString() }/{ fundingSchedule.name }
                  </Typography>
                </div>
                <div className="col-span-1 row-span-2 flex justify-end">
                  <ChevronRight className="align-middle h-full"/>
                </div>
              </div>
            </ListItem>
            <Divider/>

            <ListItem dense className="opacity-50">
              <ListItemIcon>
                <TrackChanges/>
              </ListItemIcon>
              <div className="grid grid-cols-3 grid-rows-2 grid-flow-col gap-1 w-full">
                <div className="col-span-3">
                  <Typography>
                    Contribution Option (WIP)
                  </Typography>
                </div>
                <div className="col-span-3 opacity-50">
                  <Typography variant="body2">
                    Set aside target amount
                  </Typography>
                </div>
                <div className="col-span-1 row-span-2 flex justify-end">
                  <ChevronRight className="align-middle h-full"/>
                </div>
              </div>
            </ListItem>
            <Divider/>

            <ListItem dense className="opacity-50">
              <ListItemIcon>
                <ArrowBack/>
              </ListItemIcon>
              <div className="grid grid-cols-3 grid-rows-2 grid-flow-col gap-1 w-full">
                <div className="col-span-3">
                  <Typography>
                    Money Out (WIP)
                  </Typography>
                </div>
                <div className="col-span-3 opacity-50">
                  <Typography variant="body2">
                    ....
                  </Typography>
                </div>
                <div className="col-span-1 row-span-2 flex justify-end">
                  <ChevronRight className="align-middle h-full"/>
                </div>
              </div>
            </ListItem>
          </List>
          <div className="grid grid-cols-2 grid-flow-col mb-5">
            <div className="col-span-1">
              <Button variant="outlined" color="secondary" className="w-10/12" onClick={ this.deleteExpense }>
                <DeleteOutline className="mr-2"/>
                Delete
              </Button>
            </div>
            <div className="col-span-1 flex justify-end">
              <Button variant="outlined" onClick={ this.openTransferDialog } className="w-10/12">
                <SwapHoriz className="mr-2"/>
                Transfer
              </Button>
            </div>
          </div>
        </div>
      </Fragment>
    )
  }
}

export default connect(
  state => ({
    expense: getSelectedExpense(state),
    fundingSchedules: getFundingSchedules(state),
  }),
  {
    fetchBalances,
    deleteSpending,
  }
)(ExpenseDetail);
