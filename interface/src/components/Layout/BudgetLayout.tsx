import styles from './BudgetLayout.module.scss';

export interface BudgetingLayoutProps {
  children: React.ReactNode;
}

// The budgeting sidebar used to render here; it lives in Sidebar now so the mobile drawer can move both columns as
// one panel.
export default function BudgetingLayout(props: BudgetingLayoutProps): React.JSX.Element {
  return <div className={styles.budgetLayoutRoot}>{props.children}</div>;
}
