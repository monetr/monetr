import { Fragment } from 'react';

import SettingsSidebar from '@monetr/interface/components/Layout/SettingsSidebar';

import styles from './SettingsLayout.module.scss';

export interface SettingsLayoutProps {
  children: React.ReactNode;
}

export default function SettingsLayout(props: SettingsLayoutProps): React.JSX.Element {
  return (
    <Fragment>
      <SettingsSidebar />
      <div className={styles.settingsLayoutRoot}>{props.children}</div>
    </Fragment>
  );
}
