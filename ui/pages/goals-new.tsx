import React, { Fragment } from 'react';
import { AddOutlined, SavingsOutlined } from '@mui/icons-material';

import { MBaseButton } from 'components/MButton';
import MTopNavigation from 'components/MTopNavigation';
import { useSpendingFiltered } from 'hooks/spending';
import { SpendingType } from 'models/Spending';

export default function GoalsNew(): JSX.Element {
  const { result: goals } = useSpendingFiltered(SpendingType.Goal);

  return (
    <Fragment>
      <MTopNavigation
        icon={ SavingsOutlined }
        title='Goals'
      >
        <MBaseButton color='primary' className='gap-1 py-1 px-2' onClick={ null }>
          <AddOutlined />
          New Goal
        </MBaseButton>
      </MTopNavigation>
      <div className='w-full h-full overflow-y-auto min-w-0'>
        <ul className='w-full flex flex-col gap-2 py-2'>
        </ul>
      </div>
    </Fragment>
  );
}
