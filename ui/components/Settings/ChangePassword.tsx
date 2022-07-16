import React from 'react';
import { Password } from '@mui/icons-material';
import { Button, TextField } from '@mui/material';

import { Formik, FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';
import request from 'shared/util/request';

interface ChangePasswordValues {
  currentPassword: string;
  newPassword: string;
  repeatPassword: string;
}

export default function ChangePassword(): JSX.Element {
  const { enqueueSnackbar } = useSnackbar();

  function validateInput(values: ChangePasswordValues): Partial<ChangePasswordValues> {
    const errors: Partial<ChangePasswordValues> = {};

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

  function updatePassword(values: ChangePasswordValues, helpers: FormikHelpers<ChangePasswordValues>) {
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

  const initialValues: ChangePasswordValues = {
    currentPassword: '',
    newPassword: '',
    repeatPassword: '',
  };

  return (
    <div>
      <Formik
        initialValues={ initialValues }
        validate={ validateInput }
        onSubmit={ updatePassword }
      >
        { ({
          values,
          errors,
          touched,
          handleChange,
          handleBlur,
          handleSubmit,
          isSubmitting,
          submitForm,
        }) => (
          <form onSubmit={ handleSubmit }>
            <span className="text-2xl">
              Password
            </span>
            <div className="grid lg:grid-cols-2 gap-2.5 mt-2.5">
              <div className="grid gap-2.5">
                { /*
                  This is here to suppress a Chrome warning about "you should have a username field
                  even if its hidden" thing.
                */ }
                <input id="username" name="username" autoComplete="username" className="hidden" />
                <TextField
                  id="current-password"
                  label="Current Password"
                  variant="outlined"
                  className="w-full"
                  type="password"
                  autoComplete="current-password"
                  disabled={ isSubmitting }
                  name="currentPassword"
                  error={ touched.currentPassword && !!errors.currentPassword }
                  helperText={ (touched.currentPassword && errors.currentPassword) ? errors.currentPassword : null }
                  onBlur={ handleBlur }
                  onChange={ handleChange }
                  value={ values.currentPassword }
                />
                <hr />
                <TextField
                  id="new-password"
                  label="New Password"
                  variant="outlined"
                  className="w-full"
                  type="password"
                  autoComplete="new-password"
                  disabled={ isSubmitting }
                  name="newPassword"
                  error={ touched.newPassword && !!errors.newPassword }
                  helperText={ (touched.newPassword && errors.newPassword) ? errors.newPassword : null }
                  onBlur={ handleBlur }
                  onChange={ handleChange }
                  value={ values.newPassword }
                />
                <TextField
                  id="repeat-password"
                  label="Repeat New Password"
                  variant="outlined"
                  className="w-full"
                  type="password"
                  autoComplete="new-password"
                  disabled={ isSubmitting }
                  name="repeatPassword"
                  error={ touched.repeatPassword && !!errors.repeatPassword }
                  helperText={ (touched.repeatPassword && errors.repeatPassword) ? errors.repeatPassword : null }
                  onBlur={ handleBlur }
                  onChange={ handleChange }
                  value={ values.repeatPassword }
                />
                <Button
                  disabled={ isSubmitting || (!values.currentPassword || !values.newPassword || !values.repeatPassword) }
                  onClick={ submitForm }
                  variant="contained"
                  className="mt-2.5"
                >
                  <Password className="mr-2.5" />
                  Update Password
                </Button>
              </div>
              <div className="flex items-center h-full">
                <p className="h-full opacity-70">
                  We strongly recommend you store your passwords in a <b>secure</b> password manager like 1Password.
                </p>
              </div>
            </div>
          </form>
        ) }
      </Formik>
    </div>
  );
}
