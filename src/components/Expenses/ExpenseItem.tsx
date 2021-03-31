import { Checkbox, LinearProgress, ListItem, ListItemIcon, Typography } from '@material-ui/core';
import FundingSchedule from 'data/FundingSchedule';
import Spending from 'data/Spending';
import React, { Component } from 'react';
import { connect } from 'react-redux';
import { getFundingScheduleById } from 'shared/fundingSchedules/selectors/getFundingScheduleById';
import selectExpense from 'shared/spending/actions/selectExpense';
import { getExpenseIsSelected } from 'shared/spending/selectors/getExpenseIsSelected';
import { getSpendingById } from 'shared/spending/selectors/getSpendingById';

export interface PropTypes {
  expenseId: number;
}

interface WithConnectionPropTypes extends PropTypes {
  expense: Spending;
  fundingSchedule: FundingSchedule;
  isSelected: boolean;
  selectExpense: { (expenseId: number): void };
}

export class ExpenseItem extends Component<WithConnectionPropTypes, any> {

  onClick = () => {
    return this.props.selectExpense(this.props.expenseId);
  };

  render() {
    const { expense, isSelected, fundingSchedule } = this.props;

    return (
      <ListItem button onClick={ this.onClick }>
        <ListItemIcon>
          <Checkbox
            edge="start"
            checked={ isSelected }
            tabIndex={ -1 }
            color="primary"
          />
        </ListItemIcon>
        <div className="grid grid-cols-4 grid-rows-4 grid-flow-col gap-1 w-full">
          <div className="col-span-3">
            <Typography>{ expense.name }</Typography>
          </div>
          <div className="col-span-3">
            <Typography>{ expense.getCurrentAmountString() } of { expense.getTargetAmountString() }</Typography>
          </div>
          <div className="col-span-3">
            <Typography>
              { expense.nextRecurrence.format('MMM Do') }
              { expense.description && ` - ${ expense.description }` }
            </Typography>
          </div>
          <div className="col-span-3">
            <Typography>{ expense.getNextContributionAmountString() }/{ fundingSchedule.name }</Typography>
          </div>
          <div className="col-span-1 row-span-2">
            <LinearProgress variant="determinate" color="primary"
                            value={ (expense.currentAmount / expense.targetAmount) * 100 }/>
          </div>
        </div>
      </ListItem>
    )
  }
}

export default connect(
  (state, props: PropTypes) => {
    const expense = getSpendingById(props.expenseId)(state);
    const isSelected = getExpenseIsSelected(props.expenseId)(state);

    return {
      expense,
      isSelected,
      fundingSchedule: getFundingScheduleById(expense.fundingScheduleId)(state),
    };
  },
  {
    selectExpense,
  },
)(ExpenseItem);
