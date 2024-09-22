import React from 'react';

import Card from '@monetr/interface/components/Card';
import { MBaseButton } from '@monetr/interface/components/MButton';
import MSpan from '@monetr/interface/components/MSpan';
import { showEnableTOTPModal } from '@monetr/interface/components/settings/security/EnableTOTPMModal';
import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';

import { Shield } from 'lucide-react';

export default function TOTPCard(): JSX.Element {
  const { login } = useAuthentication();

  return (
    <Card className='md:w-1/3'>
      <div className='flex justify-between items-center'>
        <div className='border-dark-monetr-border rounded border w-fit p-2 bg-dark-monetr-background-subtle'>
          <Shield />
        </div>
        <MBaseButton 
          variant='solid' 
          color='primary' 
          disabled={ Boolean(login.totpEnabledAt) }
          onClick={ showEnableTOTPModal }
        >
          { Boolean(login.totpEnabledAt) ? 'Already Enabled' : 'Enable TOTP' }
        </MBaseButton>
      </div>
      <MSpan size='md' weight='medium' color='emphasis'>
        Authenticator App (TOTP)
      </MSpan>
      <MSpan component='p'>
        Get verification codes from an authenticator app such as 1Password or Google Authenticator. It works even if
        your phone is offline.
      </MSpan>
    </Card>
  );
}
