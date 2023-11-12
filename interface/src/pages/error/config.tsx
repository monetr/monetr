import React from 'react';
import { HeartBrokenOutlined } from '@mui/icons-material';

import MSpan from '@monetr/interface/components/MSpan';

export default function ConfigError(): JSX.Element {
  return (
    <div className='w-full h-full flex items-center justify-center flex-col p-4'>
      <MSpan className='w-full h-full justify-center flex-col text-center gap-4'>
        <HeartBrokenOutlined className='text-9xl' />
        <MSpan size='xl' weight='medium'>
          There was a problem loading the monetr application config, the API may be unavailable at this time.
        </MSpan>
        <MSpan size='lg' className='gap-1'>
          You can try reloading this page, but if the problem persists please contact
          <a
            target='_blank'
            className='text-dark-monetr-blue hover:underline focus:ring-2 focus:ring-dark-monetr-blue focus:underline'
            href='mailto:support@monetr.app'
          >
            support@monetr.app
          </a>
        </MSpan>
      </MSpan>
    </div>
  );
}
