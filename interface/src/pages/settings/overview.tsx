import { Button } from '@monetr/interface/components/Button';
import MSelect from '@monetr/interface/components/MSelect';
import MTextField from '@monetr/interface/components/MTextField';
import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';

export default function SettingsOverview(): JSX.Element {
  const { data: me } = useAuthentication();

  const timezone = {
    label: me?.user?.account?.timezone,
    value: 0,
  };

  return (
    <div className='w-full h-full flex flex-col justify-between pb-4'>
      <div className='w-full flex p-4 flex-col'>
        <MTextField
          label='First Name'
          name='firstName'
          className='max-w-[24rem] w-full'
          value={me?.user?.login?.firstName}
          disabled
        />
        <MTextField
          label='Last Name'
          name='lastName'
          className='max-w-[24rem] w-full'
          value={me?.user?.login?.lastName}
          disabled
        />
        <MTextField
          label='Email Address'
          name='email'
          className='max-w-[24rem] w-full'
          value={me?.user?.login.email}
          disabled
        />
        <MSelect
          label='Timezone'
          name='timezone'
          className='max-w-[24rem] w-full'
          options={[timezone]}
          value={timezone}
          disabled
        />
      </div>
      <div className='w-full flex justify-end px-4'>
        <Button variant='primary' disabled>
          Save Settings
        </Button>
      </div>
    </div>
  );
}
