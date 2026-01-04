import { Fragment } from 'react';
import { Outlet } from 'react-router-dom';

import BudgetingSidebar from '@monetr/interface/components/Layout/BudgetingSidebar';

import styles from './BudgetLayout.module.scss';

export default function BudgetingLayout(): JSX.Element {
  // className='hidden lg:flex'
  // className='min-w-0 flex flex-col grow'
  //
  // <BudgetingSidebar />
  return (
    <Fragment>
      <div className={styles.budgetLayoutRoot}>
        <Outlet />
      </div>
    </Fragment>
  );
}
