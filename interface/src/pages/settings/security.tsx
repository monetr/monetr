import { Mail, RectangleEllipsis } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import Card from '@monetr/interface/components/Card';
import MSpan from '@monetr/interface/components/MSpan';
import { showChangePasswordModal } from '@monetr/interface/components/settings/security/ChangePasswordModal';
import TOTPCard from '@monetr/interface/components/settings/security/TOTPCard';

export default function SettingsSecurity(): JSX.Element {
  return (
    <div className='p-4 flex flex-col gap-4'>
      <div>
        <MSpan size='3xl' weight='semibold' color='emphasis' component='h1'>
          Security Settings
        </MSpan>
        <MSpan size='md' weight='normal'>
          Manage your password and multi-factor authentication.
        </MSpan>
      </div>

      <div className='mt-4 flex gap-4 flex-col md:flex-row'>
        <Card className='md:w-1/3'>
          <div className='flex justify-between items-center'>
            <div className='border-dark-monetr-border rounded border w-fit p-2 bg-dark-monetr-background-subtle'>
              <RectangleEllipsis />
            </div>
            <Button variant='primary' onClick={showChangePasswordModal}>
              Change Password
            </Button>
          </div>
          <MSpan size='md' weight='medium' color='emphasis'>
            Account Password
          </MSpan>
          <MSpan component='p'>Set a secure and unique password to make sure your account stays protected.</MSpan>
        </Card>

        <Card className='md:w-1/3'>
          <div className='flex justify-between items-center'>
            <div className='border-dark-monetr-border rounded border w-fit p-2 bg-dark-monetr-background-subtle'>
              <Mail />
            </div>
            <Button variant='primary' disabled>
              Update Email
            </Button>
          </div>
          <MSpan size='md' weight='medium' color='emphasis'>
            Email Address
          </MSpan>
          <MSpan component='p'>
            Change your primary email address, this is what you'll use to login to monetr and can be used to recover
            your acount.
          </MSpan>
        </Card>

        <TOTPCard />
      </div>
    </div>
  );
}
