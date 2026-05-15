import { Settings } from 'lucide-react';

import Divider from '@monetr/interface/components/Divider';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import { textVariants } from '@monetr/interface/components/Typography';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import { Link, useLocation } from 'wouter';

export interface SettingsLayoutProps {
  children: React.ReactNode;
}

export default function SettingsLayout(props: SettingsLayoutProps): JSX.Element {
  const config = useAppConfiguration();

  return (
    <div className='w-full h-full min-w-0 flex flex-col'>
      <MTopNavigation icon={Settings} title='Settings' />
      <div className='w-full flex px-4 mt-4 gap-6'>
        <SettingTab to='/settings/overview'>Overview</SettingTab>
        <SettingTab to='/settings/security'>Security</SettingTab>
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

function SettingTab(props: SettingTabProps): JSX.Element {
  const [pathname] = useLocation();
  const active = pathname === props.to;
  const className = mergeTailwind(
    textVariants({
      size: 'inherit',
      weight: active ? 'bold' : 'normal',
    }),
    'cursor-pointer pb-3',
  );

  return (
    <Link className={className} to={props.to}>
      {props.children}
    </Link>
  );
}
