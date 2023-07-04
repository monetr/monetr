/* eslint-disable max-len */
import React, { Fragment } from 'react';
import { MenuOutlined, PriceCheckOutlined } from '@mui/icons-material';

import ExpenseItem from './ExpenseItem';

import { useSpendingFiltered } from 'hooks/spending';
import { SpendingType } from 'models/Spending';

export default function ExpenseList(): JSX.Element {
  const { result: expenses } = useSpendingFiltered(SpendingType.Expense);

  return (
    <Fragment>
      <div className='w-full h-12 flex items-center px-4 gap-4'>
        <MenuOutlined className='visible lg:hidden text-zinc-50 cursor-pointer' />
        <span className='text-2xl text-zinc-50 font-bold flex gap-2 items-center'>
          <PriceCheckOutlined />
          Expenses
        </span>
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
