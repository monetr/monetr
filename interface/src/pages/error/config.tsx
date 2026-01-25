import { HeartCrack } from 'lucide-react';

import MSpan from '@monetr/interface/components/MSpan';
import Typography from '@monetr/interface/components/Typography';

export default function ConfigError(): JSX.Element {
  return (
    <div className='w-screen h-screen flex items-center justify-center flex-col p-4'>
      <MSpan className='w-full h-full justify-center flex-col text-center gap-4'>
        <HeartCrack className='size-24' />
        <Typography size='xl' weight='medium'>
          There was a problem loading the monetr application config, the API may be unavailable at this time.
        </Typography>
        <MSpan className='gap-1' size='lg'>
          You can try reloading this page, but if the problem persists please contact
          <a
            className='text-dark-monetr-blue hover:underline focus:ring-2 focus:ring-dark-monetr-blue focus:underline'
            href='mailto:support@monetr.app'
            rel='noopener'
            target='_blank'
          >
            support@monetr.app
          </a>
        </MSpan>
      </MSpan>
    </div>
  );
}
