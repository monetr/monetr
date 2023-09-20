import React, { Fragment } from 'react';
import { AccessTimeOutlined } from '@mui/icons-material';

import MBadge from 'components/MBadge';
import { MBaseButton } from 'components/MButton';
import MDivider from 'components/MDivider';
import MSpan from 'components/MSpan';
import { format, isThisYear } from 'date-fns';
import { useAuthenticationSink } from 'hooks/useAuthentication';

export default function SettingsBilling(): JSX.Element {
  const { result } = useAuthenticationSink();


  return (
    <div className="w-full flex flex-col p-4 max-w-xl">
      <MSpan size='2xl' weight='bold' color='emphasis' className='mb-4'>
        Billing
      </MSpan>
      <MDivider />

      <TrialingRow trialingUntil={ result?.trialingUntil } />

      <MBaseButton className='ml-auto mt-4 max-w-xs' color='primary'>
        Upgrade To A Paid Subscription
      </MBaseButton>
    </div>
  );
}

interface TrialingRowProps {
  trialingUntil: Date | null;
}

function TrialingRow(props: TrialingRowProps): JSX.Element {
  let trialEndDate: string | null;
  if (props.trialingUntil) {
    trialEndDate = isThisYear(props.trialingUntil) ?
      format(props.trialingUntil, 'MMMM do') :
      format(props.trialingUntil, 'MMMM do, yyyy');
  }

  if (!props.trialingUntil) {
    return null;
  }

  return (
    <Fragment>
      <div className='flex justify-between py-4'>
        <MSpan>
          Subscription Status
        </MSpan>
        <MBadge className='bg-yellow-600'>
          <AccessTimeOutlined />
          Trialing Until { trialEndDate }
        </MBadge>
      </div>
      <MDivider />
    </Fragment>
  );
}
