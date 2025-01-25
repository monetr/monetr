import React from 'react';
import { Save, Settings } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';

export default function BankAccountSettingsPage(): JSX.Element {
  return (
    <div className='w-full h-full flex flex-col'>
      <MTopNavigation
        icon={ Settings }
        title='Transactions'
        base='/'
        breadcrumb='Bank'
      >
        <Button variant='primary' className='gap-1 py-1 px-2' type='submit'>
          <Save />
          Save Changes
        </Button>
      </MTopNavigation>
    </div>
  );
}
