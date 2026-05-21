import { HeartCrack } from 'lucide-react';

import { layoutVariants } from '@monetr/interface/components/Layout';
import Typography from '@monetr/interface/components/Typography';

import styles from './config.module.scss';

export default function ConfigError(): JSX.Element {
  return (
    <div className={styles.root}>
      <Typography align='center' className={styles.message} size='inherit'>
        <HeartCrack className={layoutVariants({ size: 'logo' })} />
        <Typography size='xl' weight='medium'>
          There was a problem loading the monetr application config, the API may be unavailable at this time.
        </Typography>
        <Typography className={styles.contact} size='lg'>
          You can try reloading this page, but if the problem persists please contact
          <a
            className={styles.supportLink}
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
