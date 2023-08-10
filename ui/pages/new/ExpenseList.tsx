/* eslint-disable max-len */
import React, { Fragment } from 'react';
import { AddOutlined, PriceCheckOutlined } from '@mui/icons-material';

import ExpenseItem from './ExpenseItem';
import { showNewExpenseModal } from './NewExpenseModal';

import { MBaseButton } from 'components/MButton';
import MSidebarToggle from 'components/MSidebarToggle';
import { useSpendingFiltered } from 'hooks/spending';
import { SpendingType } from 'models/Spending';

export default function ExpenseList(): JSX.Element {
  const { result: expenses } = useSpendingFiltered(SpendingType.Expense);

  return (
    <Fragment>
      <div className='w-full h-12 flex items-center px-4 gap-4 justify-between'>
        <div className='flex items-center gap-4'>
          <MSidebarToggle />
          <span className='text-2xl dark:text-dark-monetr-content-emphasis font-bold flex gap-2 items-center'>
            <PriceCheckOutlined />
            Expenses
          </span>
        </div>
        <MBaseButton color='primary' className='gap-1 py-1 px-2' onClick={ showNewExpenseModal }>
          <AddOutlined />
          New Expense
        </MBaseButton>
      </div>
      <div className='w-full h-full overflow-y-auto min-w-0'>
        <ul className='w-full flex flex-col gap-2 py-2'>
          { expenses
            ?.sort((a, b) => a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1)
            .map(item => (<ExpenseItem spending={ item } key={ item.spendingId } />)) }
        </ul>
      </div>
    </Fragment>
  );
}
