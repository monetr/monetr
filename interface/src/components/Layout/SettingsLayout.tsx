import styles from './SettingsLayout.module.scss';

export interface SettingsLayoutProps {
  children: React.ReactNode;
}

export default function SettingsLayout(props: SettingsLayoutProps): React.JSX.Element {
  return <div className={styles.settingsLayoutRoot}>{props.children}</div>;
}
