import { HeartCrack } from 'lucide-react';

import { layoutVariants } from '@monetr/interface/components/Layout';
import Typography from '@monetr/interface/components/Typography';

export default function ConfigError(): JSX.Element {
  return (
    <div className='w-screen h-screen flex items-center justify-center flex-col p-4'>
      <Typography className='w-full h-full justify-center flex-col text-center gap-4' size='inherit'>
        <HeartCrack className={layoutVariants({ size: 'logo' })} />
        <Typography size='xl' weight='medium'>
          There was a problem loading the monetr application config, the API may be unavailable at this time.
        </Typography>
        <Typography className='gap-1' size='lg'>
          You can try reloading this page, but if the problem persists please contact
          <a
            className='text-dark-monetr-blue hover:underline focus:ring-2 focus:ring-dark-monetr-blue focus:underline'
            href='mailto:support@monetr.app'
            rel='noopener noreferrer'
            target='_blank'
          >
            support@monetr.app
          </a>
        </Typography>
      </Typography>
    </div>
  );
}
