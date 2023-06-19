import React, { Fragment } from 'react';
import { Divider, List } from '@mui/material';

import { showCreateExpenseDialog } from 'components/Expenses/CreateExpenseDialog';
import ExpenseItem from 'components/Expenses/ExpenseItem';
import MButton from 'components/MButton';
import { useSpendingFiltered } from 'hooks/spending';
import { SpendingType } from 'models/Spending';

export default function ExpensesNew(): JSX.Element {
  const { isLoading, result: expenses } = useSpendingFiltered(SpendingType.Expense);

  if (expenses?.length === 0) {
    return <EmptyState />;
  }

  return (
    <List disablePadding className='w-full'>
      {
        expenses
          .sort((a, b) => a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1)
          .map(expense => (
            <Fragment key={ expense.spendingId }>
              <ExpenseItem expense={ expense } />
              <Divider />
            </Fragment>
          ))
      }
    </List>
  );
}

function EmptyState(): JSX.Element {
  return (
    <div className="h-full w-full flex justify-center items-center">
      <div className="flex flex-col gap-2">
        <p className='text-3xl opacity-50'>
          You don't have any expenses yet...
        </p>
        <MButton
          onClick={ showCreateExpenseDialog }
          color="primary"
        >
          <p className='text-lg'>
            Create An Expense
          </p>
        </MButton>
      </div>
    </div>
  );
}
