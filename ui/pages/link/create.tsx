import React from 'react';
import { PowerOutlined } from '@mui/icons-material';

import MTopNavigation from 'components/MTopNavigation';
import SetupPage from 'pages/setup';

export default function LinkCreatePage(): JSX.Element {
  return (
    <div className='flex flex-col w-full'>
      <MTopNavigation
        icon={ PowerOutlined }
        title='Add another connection'
      />
      <SetupPage alreadyOnboarded manualEnabled />
    </div>
  );
}
