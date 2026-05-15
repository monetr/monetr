import { Fragment } from 'react';

import BudgetingSidebar from '@monetr/interface/components/Layout/BudgetingSidebar';

import styles from './BudgetLayout.module.scss';

export interface BudgetingLayoutProps {
  children: React.ReactNode;
}

export default function BudgetingLayout(props: BudgetingLayoutProps): JSX.Element {
  // className='hidden lg:flex'
  // className='min-w-0 flex flex-col grow'
  return (
    <Fragment>
      <BudgetingSidebar />
      <div className={styles.budgetLayoutRoot}>{props.children}</div>
    </Fragment>
  );
}
