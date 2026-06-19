import { Mail, RectangleEllipsis } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import Card from '@monetr/interface/components/Card';
import { showChangePasswordModal } from '@monetr/interface/components/settings/security/ChangePasswordModal';
import TOTPCard from '@monetr/interface/components/settings/security/TOTPCard';
import Typography from '@monetr/interface/components/Typography';

import styles from './security.module.scss';

export default function SettingsSecurity(): React.JSX.Element {
  return (
    <div className={styles.root}>
      <div>
        <Typography color='emphasis' component='h1' size='3xl' weight='semibold'>
          Security Settings
        </Typography>
        <Typography size='md' weight='normal'>
          Manage your password and multi-factor authentication.
        </Typography>
      </div>

      <div className={styles.cards}>
        <Card className={styles.card}>
          <div className={styles.header}>
            <div className={styles.iconBox}>
              <RectangleEllipsis />
            </div>
            <Button onClick={showChangePasswordModal} variant='primary'>
              Change Password
            </Button>
          </div>
          <Typography color='emphasis' size='md' weight='medium'>
            Account Password
          </Typography>
          <Typography component='p' size='inherit'>
            Set a secure and unique password to make sure your account stays protected.
          </Typography>
        </Card>

        <Card className={styles.card}>
          <div className={styles.header}>
            <div className={styles.iconBox}>
              <Mail />
            </div>
            <Button disabled variant='primary'>
              Update Email
            </Button>
          </div>
          <Typography color='emphasis' size='md' weight='medium'>
            Email Address
          </Typography>
          <Typography component='p' size='inherit'>
            Change your primary email address, this is what you&apos;ll use to login to monetr and can be used to
            recover your acount.
          </Typography>
        </Card>

        <TOTPCard />
      </div>
    </div>
  );
}
