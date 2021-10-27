import { Checkbox, Chip, LinearProgress, ListItem, ListItemIcon, Typography } from '@material-ui/core';
import FundingSchedule from 'data/FundingSchedule';
import Spending from 'data/Spending';
import { getActiveElement } from 'formik';
import moment from 'moment';
import React, { Component } from 'react';
import { connect } from 'react-redux';
import { getFundingScheduleById } from 'shared/fundingSchedules/selectors/getFundingScheduleById';
import selectExpense from 'shared/spending/actions/selectExpense';
import { getExpenseIsSelected } from 'shared/spending/selectors/getExpenseIsSelected';
import { getSpendingById } from 'shared/spending/selectors/getSpendingById';
import { AppState } from 'store';

export interface PropTypes {
  expenseId: number;
}

interface WithConnectionPropTypes extends PropTypes {
  expense: Spending;
  fundingSchedule: FundingSchedule;
  isSelected: boolean;
  selectExpense: { (expenseId: number): void };
}

export class ExpenseItem extends Component<WithConnectionPropTypes> {

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
        <div className="w-full grid grid-cols-6 grid-rows-4 grid-flow-col">
          <div className="col-span-4">
            <Typography>
              <b>{ expense.name }</b>
            </Typography>
          </div>
          <div className="col-span-4">
            <Typography
              variant="body1"
            >
              { expense.getCurrentAmountString() } <span
              className="opacity-80">of</span> { expense.getTargetAmountString() }
            </Typography>
          </div>
          <div className="col-span-4">
            <Typography
              variant="body1"
            >
              { expense.getNextOccurrenceString() }
              { expense.description && ` - ${ expense.description }` }
            </Typography>
          </div>
          <div className="col-span-4">
            <Typography
              variant="body1"
            >
              { expense.getNextContributionAmountString() }/{ fundingSchedule.name }
            </Typography>
          </div>
          <div className="flex justify-end p-5 align-middle col-span-1 row-span-4">
            { expense.isBehind &&
            <Chip
              className="self-center"
              label="Behind"
              color="secondary"
            />
            }
          </div>
          <div className="flex justify-end align-middle col-span-1 row-span-4">
            <LinearProgress
              variant="determinate"
              color="primary"
              className="self-center w-full"
              value={ Math.min((expense.currentAmount / expense.targetAmount) * 100, 100) }
            />
          </div>
        </div>
      </ListItem>
    )
  }
}

export default connect(
  (state: AppState, props: PropTypes) => {
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
