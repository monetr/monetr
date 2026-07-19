import { Settings } from 'lucide-react';
import { Link, useLocation } from 'wouter';

import Divider from '@monetr/interface/components/Divider';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import { textVariants } from '@monetr/interface/components/Typography';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './SettingsLayout.module.scss';

export interface SettingsLayoutProps {
  children: React.ReactNode;
}

export default function SettingsLayout(props: SettingsLayoutProps): React.JSX.Element {
  const config = useAppConfiguration();

  return (
    <div className={styles.root}>
      <MTopNavigation icon={Settings} title='Settings' />
      <div className={styles.tabs}>
        <SettingTab to='/settings/overview'>Overview</SettingTab>
        <SettingTab to='/settings/security'>Security</SettingTab>
        <SettingTab to='/settings/api'>API Keys</SettingTab>
        {config?.data?.billingEnabled && <SettingTab to='/settings/billing'>Billing</SettingTab>}
        <SettingTab to='/settings/about'>About</SettingTab>
      </div>
      <Divider />
      {props.children}
    </div>
  );
}

interface SettingTabProps {
  to: string;
  children: React.ReactNode;
}

function SettingTab(props: SettingTabProps): React.JSX.Element {
  const [pathname] = useLocation();
  const active = pathname === props.to;
  const className = mergeClasses(
    textVariants({
      size: 'inherit',
      weight: active ? 'bold' : 'normal',
    }),
    styles.tab,
  );

  return (
    <Link className={className} to={props.to}>
      {props.children}
    </Link>
  );
}
