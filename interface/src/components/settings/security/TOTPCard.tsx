import { Shield } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import Card from '@monetr/interface/components/Card';
import Typography from '@monetr/interface/components/Typography';
import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';

import styles from './TOTPCard.module.scss';

const showEnableTOTPModal = async () =>
  await import('@monetr/interface/components/settings/security/EnableTOTPMModal').then(modal =>
    modal.showEnableTOTPModal(),
  );

export default function TOTPCard(): JSX.Element {
  const {
    data: {
      user: { login },
    },
  } = useAuthentication();

  return (
    <Card className={styles.card}>
      <div className={styles.header}>
        <div className={styles.iconBox}>
          <Shield />
        </div>
        <Button disabled={Boolean(login.totpEnabledAt)} onClick={showEnableTOTPModal} variant='primary'>
          {login.totpEnabledAt ? 'Already Enabled' : 'Enable TOTP'}
        </Button>
      </div>
      <Typography color='emphasis' size='md' weight='medium'>
        Authenticator App (TOTP)
      </Typography>
      <Typography component='p' size='inherit'>
        Get verification codes from an authenticator app such as 1Password or Google Authenticator. It works even if
        your phone is offline.
      </Typography>
    </Card>
  );
}
