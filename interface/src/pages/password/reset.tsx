import { useEffect } from 'react';
import type { FormikErrors, FormikHelpers } from 'formik';
import { useLocation, useSearch } from 'wouter';

import FormButton from '@monetr/interface/components/FormButton';
import FormTextField from '@monetr/interface/components/FormTextField';
import MForm from '@monetr/interface/components/MForm';
import MLink from '@monetr/interface/components/MLink';
import MLogo from '@monetr/interface/components/MLogo';
import Typography from '@monetr/interface/components/Typography';
import useResetPassword from '@monetr/interface/hooks/useResetPassword';
import { useSnackbar } from '@monetr/notify';

import styles from './reset.module.scss';

interface ResetPasswordValues {
  password: string;
  verifyPassword: string;
}

const initialValues: ResetPasswordValues = {
  password: '',
  verifyPassword: '',
};

export default function PasswordResetNew(): React.JSX.Element {
  const { enqueueSnackbar } = useSnackbar();
  const [pathname, navigate] = useLocation();
  const resetPassword = useResetPassword();
  const query = new URLSearchParams(useSearch());
  // The reason indicates whether the user was forced here by a `PASSWORD_CHANGE_REQUIRED` login response (in which
  // case we show a different message) or arrived via the password reset link in their email.
  const reason = query.get('reason');
  const message =
    reason === 'password_change_required'
      ? 'You are required to change your password before authenticating.'
      : 'Enter the new password you would like to use.';
  // The token is provided as a query parameter, either from the email link or from the login flow when a password
  // change is required.
  const token = query.get('token');

  useEffect(() => {
    if (!token) {
      navigate('/login');
      enqueueSnackbar('You must get a password reset link to change your password.', {
        variant: 'warning',
        disableWindowBlurListener: true,
      });
    }

    // Clear the URL so that the token is not shown. But also so that the user cannot accidentally navigate back to the
    // password reset page with the token still in place.
    window.history.replaceState({}, document.title, !token ? '/login' : pathname);
  }, [token, enqueueSnackbar, navigate, pathname]);

  async function submit(values: ResetPasswordValues, helpers: FormikHelpers<ResetPasswordValues>): Promise<void> {
    // Without a token there is nothing to reset against. The effect above already redirects the user away in this case,
    // so just bail here to keep things type safe.
    if (!token) {
      return Promise.resolve();
    }

    helpers.setSubmitting(true);

    return await resetPassword(values.password, token)
      // If the reset password fails, then set submitting to false and do nothing. The error will already have been
      // displayed by the reset password function. We only do this if there is an error because if this succeeds then
      // the user is automatically redirected to the login page.
      .catch(() => helpers.setSubmitting(false));
  }

  function validate(values: ResetPasswordValues): FormikErrors<ResetPasswordValues> {
    const errors: FormikErrors<ResetPasswordValues> = {};

    if (values.password) {
      if (values.password.trim().length < 8) {
        errors.password = 'Password must be at least 8 characters long.';
      }
    }

    if (values.verifyPassword && values.password !== values.verifyPassword) {
      errors.verifyPassword = 'Passwords must match.';
    }

    return errors;
  }

  return (
    <MForm className={styles.root} initialValues={initialValues} onSubmit={submit} validate={validate}>
      <div className={styles.container}>
        <div className={styles.logo}>
          <MLogo />
        </div>
        <Typography className={styles.message} size='inherit'>
          {message}
        </Typography>
        <FormTextField
          autoComplete='current-password'
          autoFocus
          className={styles.input}
          label='Password'
          name='password'
          required
          type='password'
        />
        <FormTextField
          autoComplete='current-password'
          className={styles.input}
          label='Verify Password'
          name='verifyPassword'
          required
          type='password'
        />
        <FormButton className={styles.input} role='form' type='submit' variant='primary'>
          Reset Password
        </FormButton>
      </div>
      <div className={styles.signInRow}>
        <Typography color='subtle' size='sm'>
          Remembered your password?
        </Typography>
        <MLink size='sm' to='/login'>
          Sign in
        </MLink>
      </div>
    </MForm>
  );
}
