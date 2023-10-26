import React from 'react';
import { Link, Outlet, useLocation } from 'react-router-dom';
import { SettingsOutlined } from '@mui/icons-material';

import MDivider from 'components/MDivider';
import { MSpanDeriveClasses } from 'components/MSpan';
import MTopNavigation from 'components/MTopNavigation';
import { ReactElement } from 'components/types';
import { useAppConfigurationSink } from 'hooks/useAppConfiguration';

export default function SettingsLayout(): JSX.Element {
  const config = useAppConfigurationSink();

  return (
    <div className='w-full h-full min-w-0 flex flex-col'>
      <MTopNavigation
        icon={ SettingsOutlined }
        title='Settings'
      />
      <div className='w-full flex px-4 mt-4 gap-6'>
        <SettingTab to="/settings/overview">
          Overview
        </SettingTab>
        <SettingTab to="/settings/security">
          Security
        </SettingTab>
        { config?.data?.billingEnabled && (
          <SettingTab to="/settings/billing">
          Billing
          </SettingTab>
        ) }
        <SettingTab to="/settings/about">
          About
        </SettingTab>
      </div>
      <MDivider />
      <Outlet />
    </div>
  );
}

interface SettingTabProps {
  to: string;
  children: ReactElement;
}

function SettingTab(props: SettingTabProps): JSX.Element {
  const location = useLocation();
  const active = location.pathname === props.to;
  const className = MSpanDeriveClasses({
    className: 'cursor-pointer pb-3',
    weight: active ? 'bold' : 'normal',
  });

  return (
    <Link to={ props.to } className={ className }>
      { props.children }
    </Link>
  );
}
