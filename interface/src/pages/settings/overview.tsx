import { Button } from '@monetr/interface/components/Button';
import FormTextField from '@monetr/interface/components/FormTextField';
import Select from '@monetr/interface/components/Select';
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
        <FormTextField
          className='max-w-[24rem] w-full'
          disabled
          label='First Name'
          name='firstName'
          value={me?.user?.login?.firstName}
        />
        <FormTextField
          className='max-w-[24rem] w-full'
          disabled
          label='Last Name'
          name='lastName'
          value={me?.user?.login?.lastName}
        />
        <FormTextField
          className='max-w-[24rem] w-full'
          disabled
          label='Email Address'
          name='email'
          value={me?.user?.login.email}
        />
        <Select
          className='max-w-[24rem] w-full'
          disabled
          label='Timezone'
          name='timezone'
          onChange={() => {}}
          options={[timezone]}
          value={timezone}
        />
      </div>
      <div className='w-full flex justify-end px-4'>
        <Button disabled variant='primary'>
          Save Settings
        </Button>
      </div>
    </div>
  );
}
