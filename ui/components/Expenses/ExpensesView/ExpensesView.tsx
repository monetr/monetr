import React, { Fragment } from 'react';
import { Button, Divider, List, Typography } from '@mui/material';

import { showCreateExpenseDialog } from 'components/Expenses/CreateExpenseDialog';
import ExpenseDetail from 'components/Expenses/ExpenseDetail';
import ExpenseItem from 'components/Expenses/ExpenseItem';
import { useSpendingFiltered } from 'hooks/spending';
import { SpendingType } from 'models/Spending';

import 'components/Expenses/ExpensesView/styles/ExpensesView.scss';

export default function ExpensesView(): JSX.Element {
  const { result: expenses } = useSpendingFiltered(SpendingType.Expense);

  function EmptyState(): JSX.Element {
    return (
      <div className="h-full w-full bg-primary">
        <div className="view-inner h-full flex justify-center items-center">
          <div className="grid grid-cols-1 grid-rows-2 grid-flow-col gap-2">
            <Typography
              className="opacity-50"
              variant="h3"
            >
              You don't have any expenses yet...
            </Typography>
            <Button
              onClick={ showCreateExpenseDialog }
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
      </div>
    );
  }

  if (expenses.length === 0) {
    return <EmptyState />;
  }

  return (
    <div className="minus-nav bg-primary">
      <div className="flex flex-col h-full max-h-full view-inner">
        <div className="grid grid-cols-3 flex-grow">
          <div className="col-span-2">
            <div className="w-full expenses-list">
              <List disablePadding className="w-full">
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
            </div>
          </div>
          <div className="border-l">
            <div className="w-full expenses-list">
              <ExpenseDetail />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

