import { Button, Card, List, Typography } from "@material-ui/core";
import NewExpenseDialog from "components/Expenses/NewExpenseDialog";
import FundingScheduleList from "components/FundingSchedules/FundingScheduleList";
import { NewFundingScheduleDialog } from "components/FundingSchedules/NewFundingScheduleDialog";
import Spending from "data/Spending";
import { Map } from 'immutable';
import React, { Component, Fragment } from "react";
import { connect } from "react-redux";
import { getExpenses } from "shared/spending/selectors/getExpenses";

interface PropTypes {
  expenses: Map<number, Spending>;
}

interface State {
  newExpenseDialogOpen: boolean;
  showFundingSchedules: boolean;
  selectedExpense?: number;
}

export class ExpensesView extends Component<PropTypes, State> {

  state = {
    newExpenseDialogOpen: false,
    showFundingSchedules: false,
    selectedExpense: null
  };

  renderExpenseList = () => {
    const { expenses } = this.props;

    if (expenses.isEmpty()) {
      return (
        <Typography>You don't have any expenses yet...</Typography>
      )
    }

    return (
      <List disablePadding className="w-full">

      </List>
    )
  };

  renderSideBar = () => {
    const { showFundingSchedules } = this.state;
    if (showFundingSchedules) {
      return (
        <FundingScheduleList onHide={ this.hideFundingSchedules }/>
      );
    }

    return this.renderExpenseDetailView();
  };

  renderExpenseDetailView = () => {
    const { selectedExpense } = this.state;

    if (!selectedExpense) {
      return (
        <div>
          <Typography>Some stuff here</Typography>
        </div>
      );
    }

    return (
      <Typography>Selected expense: { selectedExpense }</Typography>
    );
  };

  selectExpense = (expenseId: number) => {
    return this.setState(prevState => ({
      selectedExpense: expenseId === prevState.selectedExpense ? null : expenseId
    }));
  };

  openNewExpenseDialog = () => {
    return this.setState({
      newExpenseDialogOpen: true
    });
  };

  closeNewExpenseDialog = () => {
    return this.setState({
      newExpenseDialogOpen: false
    });
  };

  showFundingSchedules = () => {
    return this.setState({
      showFundingSchedules: true
    });
  };

  hideFundingSchedules = () => {
    return this.setState({
      showFundingSchedules: false
    });
  }

  render() {
    const { newExpenseDialogOpen, showFundingSchedules } = this.state;
    return (
      <Fragment>
        <NewExpenseDialog onClose={ this.closeNewExpenseDialog } isOpen={ newExpenseDialogOpen }/>

        <div className="minus-nav">
          <div className="flex flex-col h-full p-10 max-h-full overflow-y-scroll">
            <Card elevation={ 4 } className="w-full h-13 mb-4 p-1">
              <div className="grid grid-cols-6 gap-4 flex-grow">
                <div className="col-span-4">

                </div>
                <div className="flex justify-end w-full">
                  { !showFundingSchedules &&
                  <Button className="w-full" color="secondary" onClick={ this.showFundingSchedules }>
                    Funding Schedules
                  </Button>
                  }
                </div>
                <div className="flex justify-end w-full">
                  <Button variant="outlined" className="w-full" color="primary" onClick={ this.openNewExpenseDialog }>
                    New Expense
                  </Button>
                </div>
              </div>
            </Card>
            <div className="grid grid-cols-3 gap-4 flex-grow">
              <div className="col-span-2">
                <Card elevation={ 4 } className="h-full w-full overflow-scroll table">
                  { this.renderExpenseList() }
                </Card>
              </div>
              <div className="">
                <Card elevation={ 4 } className="h-full w-full">
                  { this.renderSideBar() }
                </Card>
              </div>
            </div>
          </div>
        </div>
      </Fragment>
    );
  }
}

export default connect(
  state => ({
    expenses: getExpenses(state)
  }),
  {}
)(ExpensesView)
