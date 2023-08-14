import React from 'react';
import { PasswordOutlined } from '@mui/icons-material';

import MFormButton from 'components/MButton';
import MForm from 'components/MForm';
import MTextField from 'components/MTextField';

interface ChangePasswordValues {
  currentPassword: string;
  newPassword: string;
  repeatPassword: string;
}

const initialValues: ChangePasswordValues = {
  currentPassword: '',
  newPassword: '',
  repeatPassword: '',
};

export default function SettingsSecurity(): JSX.Element {

  function updatePassword() {

  }

  return (
    <div className="max-w-xl w-full flex p-4">
      <MForm
        onSubmit={ updatePassword }
        initialValues={ initialValues }
        className="flex flex-col w-full"
      >
        <MTextField
          autoComplete="current-password"
          className="w-full"
          label="Current Password"
          name="currentPassword"
          type="password"
        />
        <MTextField
          autoComplete="new-password"
          className="w-full"
          label="New Password"
          name="newPassword"
          type="password"
        />
        <MTextField
          autoComplete="new-password"
          className="w-full"
          label="Repeat Password"
          name="repeatPassword"
          type="password"
        />
        <MFormButton
          type="submit"
          color="primary"
          className="gap-2"
        >
          <PasswordOutlined />
          Update Password
        </MFormButton>
      </MForm>
    </div>
  );
}
