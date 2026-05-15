import styles from './SidebarPaddingLayout.module.scss';

export interface SidebarPaddingLayoutProps {
  children: React.ReactNode;
}

export default function SidebarPaddingLayout(props: SidebarPaddingLayoutProps): React.JSX.Element {
  return <div className={styles.sidebarPaddingLayout}>{props.children}</div>;
}
