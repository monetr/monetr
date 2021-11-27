import { Button, Card, Divider, List, Typography } from '@mui/material';
import ExpenseDetail from 'components/Expenses/ExpenseDetail';
import ExpenseItem from 'components/Expenses/ExpenseItem';
import NewExpenseDialog from 'components/Expenses/NewExpenseDialog';
import React, { Component, Fragment } from 'react';
import { connect } from 'react-redux';
import { getExpenseIds } from 'shared/spending/selectors/getExpenseIds';

import './styles/ExpensesView.scss';

interface WithConnectionPropTypes {
  expenseIds: number[],
}

interface State {
  newExpenseDialogOpen: boolean;
  showFundingSchedules: boolean;
}

export class ExpensesView extends Component<WithConnectionPropTypes, State> {

  state = {
    newExpenseDialogOpen: false,
    showFundingSchedules: false,
  };

  componentDidMount() {
    console.log('expenses');
  }

  renderExpenseList = () => {
    const { expenseIds } = this.props;

    if (expenseIds.length === 0) {
      return (
        <div className="h-full flex justify-center items-center">
          <div className="grid grid-cols-1 grid-rows-2 grid-flow-col gap-2">
            <Typography
              className="opacity-50"
              variant="h3"
            >
              You don't have any expenses yet...
            </Typography>
            <Button
              onClick={ this.openNewExpenseDialog }
              color="primary"
            >
              <Typography
                variant="h6"
              >
                Create An Expense
              </Typography>
            </Button>
          </div>
        </div>
      )
    }

    return (
      <List disablePadding className="w-full">
        {
          expenseIds.map(expense => (
            <Fragment key={ expense }>
              <ExpenseItem expenseId={ expense }/>
              <Divider/>
            </Fragment>
          ))
        }
      </List>
    )
  };

  renderSideBar = () => {
    return (
      <ExpenseDetail/>
    );
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
    const { expenseIds } = this.props;

    if (expenseIds.length === 0) {
      return (
        <Fragment>
          { newExpenseDialogOpen &&
          <NewExpenseDialog onClose={ this.closeNewExpenseDialog } isOpen/>
          }

          <div className="minus-nav">
            <div className="flex flex-col h-full p-10 max-h-full">
              <div className="grid grid-cols-3 gap-4 flex-grow">
                <div className="col-span-3">
                  <Card elevation={ 4 } className="w-full expenses-list">
                    { this.renderExpenseList() }
                  </Card>
                </div>
              </div>
            </div>
          </div>
        </Fragment>
      )
    }

    return (
      <Fragment>
        { newExpenseDialogOpen &&
        <NewExpenseDialog onClose={ this.closeNewExpenseDialog } isOpen/>
        }

        <div className="minus-nav">
          <div className="flex flex-col h-full p-10 max-h-full">
            <div className="grid grid-cols-3 gap-4 flex-grow">
              <div className="col-span-2">
                <Card elevation={ 4 } className="w-full expenses-list">
                  { this.renderExpenseList() }
                </Card>
              </div>
              <div className="">
                <Card elevation={ 4 } className="w-full expenses-list">
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
    expenseIds: getExpenseIds(state),
  }),
  {}
)(ExpensesView)
