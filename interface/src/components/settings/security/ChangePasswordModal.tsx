import type React from 'react';
import { useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import type { FormikErrors, FormikHelpers } from 'formik';
import { RectangleEllipsis } from 'lucide-react';
import { useSnackbar } from 'notistack';

import { Button } from '@monetr/interface/components/Button';
import FormButton from '@monetr/interface/components/FormButton';
import MForm from '@monetr/interface/components/MForm';
import MModal, { type MModalRef } from '@monetr/interface/components/MModal';
import MSpan from '@monetr/interface/components/MSpan';
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

function ChangePasswordModal(): JSX.Element {
  const modal = useModal();
  const { enqueueSnackbar } = useSnackbar();
  const ref = useRef<MModalRef>(null);

  async function updatePassword(values: ChangePasswordValues, helpers: FormikHelpers<ChangePasswordValues>) {
    helpers.setSubmitting(true);
    return request()
      .put('/users/security/password', {
        currentPassword: values.currentPassword,
        newPassword: values.newPassword,
      })
      .then(() =>
        enqueueSnackbar('Successfully updated password.', {
          variant: 'success',
          disableWindowBlurListener: true,
        }),
      )
      .then(() => modal.remove())
      .catch(error =>
        enqueueSnackbar(error?.response?.data?.error || 'Failed to change password.', {
          variant: 'error',
          disableWindowBlurListener: true,
        }),
      )
      .finally(() => helpers.setSubmitting(false));
  }

  function validate(values: ChangePasswordValues): FormikErrors<ChangePasswordValues> {
    const errors: FormikErrors<ChangePasswordValues> = {};

    if (!values.currentPassword) {
      errors.currentPassword = 'Your current password must be provided in order to change your password.';
      return errors;
    }

    if (values.newPassword.length < 8) {
      errors.newPassword = 'New Password must be at least 8 characters long.';
    }

    if (values.repeatPassword.length === 0) {
      errors.repeatPassword = 'You must repeat your password.';
    }

    if (values.repeatPassword !== values.newPassword) {
      errors.repeatPassword = 'New Passwords must match.';
    }

    return errors;
  }

  return (
    <MModal open={modal.visible} ref={ref} className='sm:max-w-sm'>
      <MForm
        onSubmit={updatePassword}
        initialValues={initialValues}
        validate={validate}
        className='h-full flex flex-col gap-2 p-2 justify-between'
      >
        <div className='flex flex-col'>
          <MSpan weight='bold' size='xl' className='mb-2'>
            Change Your Password
          </MSpan>
          <MTextField
            autoFocus
            autoComplete='current-password'
            className='w-full'
            label='Current Password'
            name='currentPassword'
            type='password'
            placeholder='********'
          />
          <MTextField
            autoComplete='new-password'
            className='w-full'
            label='New Password'
            name='newPassword'
            type='password'
            placeholder='********'
          />
          <MTextField
            autoComplete='new-password'
            className='w-full'
            label='Repeat Password'
            name='repeatPassword'
            type='password'
            placeholder='********'
          />
        </div>
        <div className='flex justify-end gap-2'>
          <Button variant='secondary' onClick={modal.remove} data-testid='close-change-password-modal'>
            Cancel
          </Button>
          <FormButton color='primary' type='submit'>
            <RectangleEllipsis className='mr-1' />
            Update Password
          </FormButton>
        </div>
      </MForm>
    </MModal>
  );
}

const changePasswordModal = NiceModal.create(ChangePasswordModal);

export default changePasswordModal;

export function showChangePasswordModal(): Promise<void> {
  return NiceModal.show<void, React.ComponentProps<typeof changePasswordModal>, {}>(changePasswordModal);
}
