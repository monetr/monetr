import { Button } from '@monetr/interface/components/Button';
import FormTextField from '@monetr/interface/components/FormTextField';
import Select from '@monetr/interface/components/Select';
import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';

import styles from './overview.module.scss';

export default function SettingsOverview(): React.JSX.Element {
  const { data: me } = useAuthentication();

  const timezone = {
    label: me?.user?.account?.timezone ?? '',
    value: 0,
  };

  return (
    <div className={styles.root}>
      <div className={styles.fields}>
        <FormTextField
          className={styles.field}
          disabled
          label='First Name'
          name='firstName'
          value={me?.user?.login?.firstName}
        />
        <FormTextField
          className={styles.field}
          disabled
          label='Last Name'
          name='lastName'
          value={me?.user?.login?.lastName}
        />
        <FormTextField
          className={styles.field}
          disabled
          label='Email Address'
          name='email'
          value={me?.user?.login.email}
        />
        <Select
          className={styles.field}
          disabled
          label='Timezone'
          name='timezone'
          onChange={() => {}}
          options={[timezone]}
          value={timezone}
        />
      </div>
      <div className={styles.actions}>
        <Button disabled variant='primary'>
          Save Settings
        </Button>
      </div>
    </div>
  );
}
