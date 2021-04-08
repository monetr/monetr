import { Button, Divider, LinearProgress, List, ListItem, ListItemIcon, Typography } from '@material-ui/core';
import { AccountBalance, ArrowBack, ArrowForward, ChevronRight, Event, TrackChanges } from '@material-ui/icons';
import TransferDialog from 'components/Spending/TransferDialog';
import FundingSchedule from 'data/FundingSchedule';
import Spending from 'data/Spending';
import { Map } from 'immutable';
import React, { Component, Fragment } from 'react';
import { connect } from 'react-redux';
import { getFundingSchedules } from 'shared/fundingSchedules/selectors/getFundingSchedules';
import { getSelectedExpense } from 'shared/spending/selectors/getSelectedExpense';

interface WithConnectionPropTypes {
  expense?: Spending;
  fundingSchedules: Map<number, FundingSchedule>;
}

interface State {
  transferDialogOpen: boolean;
}

export class ExpenseDetail extends Component<WithConnectionPropTypes, State> {

  state = {
    transferDialogOpen: false,
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


  render() {
    const { expense } = this.props;
    if (!expense) {
      return null;
    }

    const fundingSchedule = this.props.fundingSchedules.get(expense.fundingScheduleId, new FundingSchedule());

    const { transferDialogOpen } = this.state;

    return (
      <Fragment>
        { transferDialogOpen &&
        <TransferDialog isOpen onClose={ this.closeTransferDialog } toSpendingId={ expense.spendingId }/>
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
                { expense.nextRecurrence.format('MMM Do') } - { expense.description }
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
            <ListItem button dense>
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
                  <Typography>
                    { expense.getTargetAmountString() }
                  </Typography>
                </div>
                <div className="col-span-1 row-span-2 flex justify-end">
                  <ChevronRight className="align-middle h-full"/>
                </div>
              </div>
            </ListItem>
            <Divider/>

            <ListItem button dense>
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
                  <Typography>
                    { expense.description }
                  </Typography>
                </div>
                <div className="col-span-1 row-span-2 flex justify-end">
                  <ChevronRight className="align-middle h-full"/>
                </div>
              </div>
            </ListItem>
            <Divider/>

            <ListItem button dense>
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
                  <Typography>
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
                  <Typography>
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
                  <Typography>
                    ....
                  </Typography>
                </div>
                <div className="col-span-1 row-span-2 flex justify-end">
                  <ChevronRight className="align-middle h-full"/>
                </div>
              </div>
            </ListItem>
          </List>
          <div className="flex justify-center mb-5">
            <Button variant="outlined" onClick={ this.openTransferDialog }>
              Transfer
            </Button>
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
  {}
)(ExpenseDetail);
