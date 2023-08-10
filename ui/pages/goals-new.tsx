import React, { Fragment } from 'react';
import { AddOutlined, PriceCheckOutlined, SavingsOutlined } from '@mui/icons-material';


import { MBaseButton } from 'components/MButton';
import MSidebarToggle from 'components/MSidebarToggle';
import { useSpendingFiltered } from 'hooks/spending';
import { SpendingType } from 'models/Spending';

export default function GoalsNew(): JSX.Element {
  const { result: goals } = useSpendingFiltered(SpendingType.Goal);

  return (
    <Fragment>
      <div className='w-full h-12 flex items-center px-4 gap-4 justify-between'>
        <div className='flex items-center gap-4'>
          <MSidebarToggle />
          <span className='text-2xl dark:text-dark-monetr-content-emphasis font-bold flex gap-2 items-center'>
            <SavingsOutlined />
            Goals
          </span>
        </div>
        <MBaseButton color='primary' className='gap-1 py-1 px-2' onClick={ null }>
          <AddOutlined />
          New Goal
        </MBaseButton>
      </div>
      <div className='w-full h-full overflow-y-auto min-w-0'>
        <ul className='w-full flex flex-col gap-2 py-2'>
        </ul>
      </div>
    </Fragment>
  );
}
