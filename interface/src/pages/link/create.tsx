import React from 'react';
import { PowerOutlined } from '@mui/icons-material';

import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import SetupPage from '@monetr/interface/pages/setup';

export default function LinkCreatePage(): JSX.Element {
  const { data: config } = useAppConfiguration();

  return (
    <div className='flex flex-col w-full'>
      <MTopNavigation icon={PowerOutlined} title='Add another connection' />
      <SetupPage alreadyOnboarded manualEnabled={config?.manualEnabled} />
    </div>
  );
}
