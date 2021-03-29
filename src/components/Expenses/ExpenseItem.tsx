import { Checkbox, LinearProgress, ListItem, ListItemIcon, Typography } from '@material-ui/core';
import FundingSchedule from 'data/FundingSchedule';
import Spending from 'data/Spending';
import React, { Component } from 'react';
import { connect } from 'react-redux';
import { getFundingScheduleById } from 'shared/fundingSchedules/selectors/getFundingScheduleById';
import { getSpendingById } from 'shared/spending/selectors/getSpendingById';

export interface PropTypes {
  expenseId: number;
}

interface WithConnectionPropTypes extends PropTypes {
  expense: Spending;
  fundingSchedule: FundingSchedule;
}

export class ExpenseItem extends Component<WithConnectionPropTypes, any> {

  render() {
    const { expense, fundingSchedule } = this.props;

    return (
      <ListItem button>
        <ListItemIcon>
          <Checkbox
            edge="start"
            checked={ false }
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

    return {
      expense,
      fundingSchedule: getFundingScheduleById(expense.fundingScheduleId)(state),
    };
  },
  {},
)(ExpenseItem);
