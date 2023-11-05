import React, { useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { FormikErrors, FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import MFormButton from '@monetr/interface/components/MButton';
import MForm from '@monetr/interface/components/MForm';
import MLink from '@monetr/interface/components/MLink';
import MLogo from '@monetr/interface/components/MLogo';
import MSpan from '@monetr/interface/components/MSpan';
import MTextField from '@monetr/interface/components/MTextField';
import useResetPassword from '@monetr/interface/hooks/useResetPassword';

interface ResetPasswordValues {
  password: string;
  verifyPassword: string;
}

const initialValues: ResetPasswordValues = {
  password: '',
  verifyPassword: '',
};

export default function PasswordResetNew(): JSX.Element {
  const { enqueueSnackbar } = useSnackbar();
  const location = useLocation();
  const navigate = useNavigate();
  const resetPassword = useResetPassword();
  const { state: routeState } = useLocation();
  const message = (routeState && routeState['message']) || 'Enter the new password you would like to use.';
  const search = location.search;
  const query = new URLSearchParams(search);
  // The token is loaded from the route state (which is provided when a password reset is being forced) or from the
  // URL query parameter (which is provided when the user is brought here from a link in their email).
  const token = query.get('token') || (routeState && routeState['token']);

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
    window.history.replaceState({}, document.title, !token ? '/login' : location.pathname);
  }, [token, enqueueSnackbar, navigate, location.pathname]);

  async function submit(values: ResetPasswordValues, helpers: FormikHelpers<ResetPasswordValues>): Promise<void> {
    helpers.setSubmitting(true);

    return resetPassword(values.password, token)
      // If the reset password fails, then set submitting to false and do nothing. The error will already have been
      // displayed by the reset password function. We only do this if there is an error because if this succeeds then
      // the user is automatically redirected to the login page.
      .catch(() => helpers.setSubmitting(false));
  }

  function validate(values: ResetPasswordValues): FormikErrors<ResetPasswordValues> {
    const errors: FormikErrors<ResetPasswordValues> = {};

    if (values.password) {
      if (values.password.trim().length < 8) {
        errors['password'] = 'Password must be at least 8 characters long.';
      }
    }

    if (values.verifyPassword && values.password !== values.verifyPassword) {
      errors['verifyPassword'] = 'Passwords must match.';
    }

    return errors;
  }

  return (
    <MForm
      onSubmit={ submit }
      initialValues={ initialValues }
      validate={ validate }
      className="w-full h-full flex flex-col pt-10 md:pt-0 mb:pb-10 md:justify-center items-center px-5 gap-1"
    >
      <div className='flex items-center flex-col gap-1 w-full xl:w-1/5 lg:w-1/4 md:w-1/3 sm:w-1/2'>
        <div className="max-w-[128px] w-full">
          <MLogo />
        </div>
        <MSpan className='flex items-center text-center'>
          { message }
        </MSpan>
        <MTextField
          autoFocus
          autoComplete='current-password'
          label="Password"
          name='password'
          type='password'
          required
          className="w-full"
        />
        <MTextField
          autoComplete='current-password'
          label="Verify Password"
          name='verifyPassword'
          type='password'
          required
          className="w-full"
        />
        <MFormButton
          color="primary"
          variant="solid"
          role="form"
          type="submit"
          className='w-full'
        >
          Reset Password
        </MFormButton>
      </div>
      <div className="w-full lg:w-1/4 sm:w-1/3 mt-1 flex justify-center gap-1">
        <MSpan color="subtle" className='text-sm'>Remembered your password?</MSpan>
        <MLink to="/login" size="sm">Sign in</MLink>
      </div>
    </MForm>
  );
}
