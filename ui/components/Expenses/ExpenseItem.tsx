import { Checkbox, Chip, LinearProgress, ListItem, ListItemIcon, Typography } from '@mui/material';
import React from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { getFundingScheduleById } from 'shared/fundingSchedules/selectors/getFundingScheduleById';
import selectExpense from 'shared/spending/actions/selectExpense';
import { getExpenseIsSelected } from 'shared/spending/selectors/getExpenseIsSelected';
import { getSpendingById } from 'shared/spending/selectors/getSpendingById';

export interface PropTypes {
  expenseId: number;
}

const ExpenseItem = (props: PropTypes): JSX.Element => {
  const expense = useSelector(getSpendingById(props.expenseId));
  const isSelected = useSelector(getExpenseIsSelected(props.expenseId));
  const fundingSchedule = useSelector(getFundingScheduleById(expense.fundingScheduleId));

  const dispatch = useDispatch();

  const onClick = () => dispatch(selectExpense(expense.spendingId));

  return (
    <ListItem button onClick={ onClick }>
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

export default ExpenseItem;