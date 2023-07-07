import React from 'react';
import { HeartBrokenOutlined } from '@mui/icons-material';

import MSpan from 'components/MSpan';

export default function ConfigError(): JSX.Element {
  return (
    <div className="w-full h-full flex items-center justify-center flex-col p-4">
      <MSpan className="w-full h-full flex items-center justify-center flex-col text-center gap-4">
        <HeartBrokenOutlined className='text-9xl' />
        <MSpan className="text-xl font-medium">
          There was a problem loading the monetr application config, the API may be unavailable at this time.
        </MSpan>
        <MSpan className="text-lg">
          You can try reloading this page, but if the problem persists please contact support@monetr.app
        </MSpan>
      </MSpan>
    </div>
  );
}
