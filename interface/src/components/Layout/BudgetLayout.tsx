import { Fragment } from 'react';

import BudgetingSidebar from '@monetr/interface/components/Layout/BudgetingSidebar';

import styles from './BudgetLayout.module.scss';

export interface BudgetingLayoutProps {
  children: React.ReactNode;
}

export default function BudgetingLayout(props: BudgetingLayoutProps): React.JSX.Element {
  return (
    <Fragment>
      <BudgetingSidebar />
      <div className={styles.budgetLayoutRoot}>{props.children}</div>
    </Fragment>
  );
}
