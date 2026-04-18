import { Mail, RectangleEllipsis } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import Card from '@monetr/interface/components/Card';
import { showChangePasswordModal } from '@monetr/interface/components/settings/security/ChangePasswordModal';
import TOTPCard from '@monetr/interface/components/settings/security/TOTPCard';
import Typography from '@monetr/interface/components/Typography';

export default function SettingsSecurity(): JSX.Element {
  return (
    <div className='p-4 flex flex-col gap-4'>
      <div>
        <Typography color='emphasis' component='h1' size='3xl' weight='semibold'>
          Security Settings
        </Typography>
        <Typography size='md' weight='normal'>
          Manage your password and multi-factor authentication.
        </Typography>
      </div>

      <div className='mt-4 flex gap-4 flex-col md:flex-row'>
        <Card className='md:w-1/3'>
          <div className='flex justify-between items-center'>
            <div className='border-dark-monetr-border rounded border w-fit p-2 bg-dark-monetr-background-subtle'>
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

        <Card className='md:w-1/3'>
          <div className='flex justify-between items-center'>
            <div className='border-dark-monetr-border rounded border w-fit p-2 bg-dark-monetr-background-subtle'>
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
            Change your primary email address, this is what you'll use to login to monetr and can be used to recover
            your acount.
          </Typography>
        </Card>

        <TOTPCard />
      </div>
    </div>
  );
}
