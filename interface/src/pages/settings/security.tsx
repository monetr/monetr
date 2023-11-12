import React from 'react';
import { PasswordOutlined } from '@mui/icons-material';
import { FormikErrors, FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import MFormButton from '@monetr/interface/components/MButton';
import MForm from '@monetr/interface/components/MForm';
import MTextField from '@monetr/interface/components/MTextField';
import request from '@monetr/interface/util/request';

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
  const { enqueueSnackbar } = useSnackbar();

  async function updatePassword(values: ChangePasswordValues, helpers: FormikHelpers<ChangePasswordValues>) {
    helpers.setSubmitting(true);
    return request().put('/users/security/password', {
      currentPassword: values.currentPassword,
      newPassword: values.newPassword,
    })
      .then(() => enqueueSnackbar('Successfully updated password.', {
        variant: 'success',
        disableWindowBlurListener: true,
      }))
      .then(() => helpers.resetForm())
      .catch(error => enqueueSnackbar(error?.response?.data?.error || 'Failed to change password.', {
        variant: 'error',
        disableWindowBlurListener: true,
      }))
      .finally(() => helpers.setSubmitting(false));
  }

  function validate(values: ChangePasswordValues): FormikErrors<ChangePasswordValues> {
    const errors: FormikErrors<ChangePasswordValues> = {};

    if (!values.currentPassword) {
      errors['currentPassword'] = 'Your current password must be provided in order to change your password.';
      return errors;
    }

    if (values.newPassword.length < 8) {
      errors['newPassword'] = 'New Password must be at least 8 characters long.';
    }

    if (values.repeatPassword !== values.newPassword) {
      errors['repeatPassword'] = 'New Passwords must match.';
    }

    return errors;
  }

  return (
    <MForm
      onSubmit={ updatePassword }
      initialValues={ initialValues }
      validate={ validate }
      className='max-w-xl w-full flex flex-col p-4'
    >
      <MTextField
        autoComplete='current-password'
        className='w-full'
        label='Current Password'
        name='currentPassword'
        type='password'
      />
      <MTextField
        autoComplete='new-password'
        className='w-full'
        label='New Password'
        name='newPassword'
        type='password'
      />
      <MTextField
        autoComplete='new-password'
        className='w-full'
        label='Repeat Password'
        name='repeatPassword'
        type='password'
      />
      <MFormButton
        type='submit'
        color='primary'
        className='gap-2'
      >
        <PasswordOutlined />
        Update Password
      </MFormButton>
    </MForm>
  );
}
