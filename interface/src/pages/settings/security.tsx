import React from 'react';
import { PasswordOutlined } from '@mui/icons-material';
import { FormikErrors, FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import Card from '@monetr/interface/components/Card';
import MFormButton, { MBaseButton } from '@monetr/interface/components/MButton';
import MForm from '@monetr/interface/components/MForm';
import MSpan from '@monetr/interface/components/MSpan';
import MTextField from '@monetr/interface/components/MTextField';
import { showChangePasswordModal } from '@monetr/interface/components/settings/security/ChangePasswordModal';
import request from '@monetr/interface/util/request';

import { Mail, RectangleEllipsis, Shield } from 'lucide-react';

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
            <MBaseButton 
              variant='solid' 
              color='primary'
              onClick={ showChangePasswordModal }
            >
              Change Password
            </MBaseButton>
          </div>
          <MSpan size='md' weight='medium' color='emphasis'>
            Account Password
          </MSpan>
          <MSpan component='p'>
            Set a secure and unique password to make sure your account stays protected.
          </MSpan>
        </Card>

        <Card className='md:w-1/3'>
          <div className='flex justify-between items-center'>
            <div className='border-dark-monetr-border rounded border w-fit p-2 bg-dark-monetr-background-subtle'>
              <Mail />
            </div>
            <MBaseButton variant='solid' color='primary' disabled>
              Update Email
            </MBaseButton>
          </div>
          <MSpan size='md' weight='medium' color='emphasis'>
            Email Address
          </MSpan>
          <MSpan component='p'>
            Change your primary email address, this is what you'll use to login to monetr and can be used to recover
            your acount.
          </MSpan>
        </Card>

        <Card className='md:w-1/3'>
          <div className='flex justify-between items-center'>
            <div className='border-dark-monetr-border rounded border w-fit p-2 bg-dark-monetr-background-subtle'>
              <Shield />
            </div>
            <MBaseButton variant='solid' color='primary' disabled>
              Enable TOTP
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
      </div>
    </div>
  );

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
