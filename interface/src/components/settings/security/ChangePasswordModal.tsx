import type React from 'react';
import { useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import type { FormikErrors, FormikHelpers } from 'formik';
import { RectangleEllipsis } from 'lucide-react';
import { useSnackbar } from 'notistack';

import { Button } from '@monetr/interface/components/Button';
import FormButton from '@monetr/interface/components/FormButton';
import FormTextField from '@monetr/interface/components/FormTextField';
import MForm from '@monetr/interface/components/MForm';
import MModal, { type MModalRef } from '@monetr/interface/components/MModal';
import MSpan from '@monetr/interface/components/MSpan';
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
    <MModal className='sm:max-w-sm' open={modal.visible} ref={ref}>
      <MForm
        className='h-full flex flex-col gap-2 p-2 justify-between'
        initialValues={initialValues}
        onSubmit={updatePassword}
        validate={validate}
      >
        <div className='flex flex-col'>
          <MSpan className='mb-2' size='xl' weight='bold'>
            Change Your Password
          </MSpan>
          <FormTextField
            autoComplete='current-password'
            autoFocus
            className='w-full'
            label='Current Password'
            name='currentPassword'
            placeholder='********'
            type='password'
          />
          <FormTextField
            autoComplete='new-password'
            className='w-full'
            label='New Password'
            name='newPassword'
            placeholder='********'
            type='password'
          />
          <FormTextField
            autoComplete='new-password'
            className='w-full'
            label='Repeat Password'
            name='repeatPassword'
            placeholder='********'
            type='password'
          />
        </div>
        <div className='flex justify-end gap-2'>
          <Button data-testid='close-change-password-modal' onClick={modal.remove} variant='secondary'>
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
  return NiceModal.show<void, React.ComponentProps<typeof changePasswordModal>, unknown>(changePasswordModal);
}
