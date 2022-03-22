import { Password } from '@mui/icons-material';
import { Button, TextField } from '@mui/material';
import { FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';
import React from 'react';

interface ChangePasswordValues {
  currentPassword: string;
  newPassword: string;
  repeatPassword: string;
}

export default function ChangePassword(): JSX.Element {
  const { enqueueSnackbar } = useSnackbar();

  function validateInput(values: ChangePasswordValues): Partial<ChangePasswordValues> {
    let errors: Partial<ChangePasswordValues> = {};

    if (!values.currentPassword) {
      errors['currentPassword'] = 'Your current password must be provided in order to change your password.';
    }

    if (values.newPassword.length < 8) {
      errors['newPassword'] = 'New Password must be at least 8 characters long.'
    }

    if (values.repeatPassword !== values.newPassword) {
      errors['repeatPassword'] = 'New Passwords must match.'
    }

    return errors;
  }

  function updatePassword(values: ChangePasswordValues, helpers: FormikHelpers<ChangePasswordValues>): Promise<void> {
    return Promise.resolve();
  }

  return (
    <div>
      <span className="text-2xl">
        Password
      </span>
      <div className="grid lg:grid-cols-2 gap-2.5 mt-2.5">
        <div className="grid gap-2.5">
          <TextField id="outlined-basic" label="Current Password" variant="outlined" className="w-full"/>
          <hr/>
          <TextField id="outlined-basic" label="New Password" variant="outlined" className="w-full"/>
          <TextField id="outlined-basic" label="Repeat New Password" variant="outlined" className="w-full"/>
          <Button variant="contained" className="mt-2.5">
            <Password className="mr-2.5"/>
            Update Password
          </Button>
        </div>
        <div className="h-full flex items-center">
          <p className="opacity-70 h-full">
            We strongly recommend you store your passwords in a <b>secure</b> password manager like 1Password.
          </p>
        </div>
      </div>
    </div>
  )
}
