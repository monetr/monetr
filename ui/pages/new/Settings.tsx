/* eslint-disable max-len */
import React from 'react';

import { MBaseButton } from 'components/MButton';
import MDivider from 'components/MDivider';
import MSelect from 'components/MSelect';
import MSpan from 'components/MSpan';
import MTextField from 'components/MTextField';
import { useAuthenticationSink } from 'hooks/useAuthentication';

export default function Settings(): JSX.Element {
  const { result: me } = useAuthenticationSink();

  const timezone = {
    label: me?.user?.account?.timezone,
    value: 0,
  };

  return (
    <div className='flex flex-col w-full py-4 h-full relative'>
      <MSpan className='mx-4 text-5xl font-medium'>
        Settings
      </MSpan>
      <div className='w-full flex px-4 mt-4 gap-6'>
        <MSpan className='cursor-pointer dark:hover:text-dark-monetr-content-emphasis font-bold dark:text-dark-monetr-brand-faint'>
          Overview
        </MSpan>
        <MSpan className='cursor-pointer dark:hover:text-dark-monetr-content-emphasis font-normal'>
          Security
        </MSpan>
        <MSpan className='cursor-pointer dark:hover:text-dark-monetr-content-emphasis font-normal'>
          About
        </MSpan>
      </div>
      <MDivider className='mt-3' />
      <div className='w-full h-full flex flex-col justify-between'>
        <div className='w-full flex p-4 flex-col'>
          <MTextField
            label='First Name'
            name='firstName'
            className='max-w-[24rem] w-full'
            value={ me?.user?.login?.firstName }
            disabled
          />
          <MTextField
            label='Last Name'
            name='lastName'
            className='max-w-[24rem] w-full'
            value={ me?.user?.login?.lastName }
            disabled
          />
          <MTextField
            label='Email Address'
            name='email'
            className='max-w-[24rem] w-full'
            value={ me?.user?.login.email }
            disabled
          />
          <MSelect
            label='Timezone'
            name='timezone'
            className='max-w-[24rem] w-full'
            options={ [timezone] }
            value={ timezone }
            disabled
          />
        </div>
        <div className='w-full flex justify-end px-4'>
          <MBaseButton color='primary'>
            Save Settings
          </MBaseButton>
        </div>
      </div>
    </div>
  );
}
