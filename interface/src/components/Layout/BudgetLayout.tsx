import styles from './BudgetLayout.module.scss';

export interface BudgetingLayoutProps {
  children: React.ReactNode;
}

export default function BudgetingLayout(props: BudgetingLayoutProps): React.JSX.Element {
  return <div className={styles.budgetLayoutRoot}>{props.children}</div>;
}
