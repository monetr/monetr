import React from 'react';
import { Checkbox, Chip, LinearProgress, ListItem, ListItemIcon, Typography } from '@mui/material';

import { useFundingSchedule } from 'hooks/fundingSchedules';
import useStore from 'hooks/store';
import Spending from 'models/Spending';

export interface Props {
  expense: Spending;
}

export default function ExpenseItem(props: Props): JSX.Element {
  const { expense } = props;

  const {
    selectedExpenseId,
    setCurrentExpense,
  } = useStore();

  const isSelected = expense.spendingId === selectedExpenseId;
  const fundingSchedule = useFundingSchedule(expense.fundingScheduleId);
  const onClick = () => setCurrentExpense(isSelected ? null : expense.spendingId);

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
            { expense.getNextContributionAmountString() }/{ fundingSchedule?.name }
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
  );
};
