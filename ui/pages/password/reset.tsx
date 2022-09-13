import React, { Fragment, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { Button, CircularProgress, TextField } from '@mui/material';
import classnames from 'classnames';
import { Formik, FormikErrors, FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import BackToLoginButton from 'components/Authentication/BackToLoginButton';
import CenteredLogo from 'components/Logo/CenteredLogo';
import useResetPassword from 'hooks/useResetPassword';

interface ResetPasswordValues {
  password: string;
  verifyPassword: string;
}

const initialValues: ResetPasswordValues = {
  password: '',
  verifyPassword: '',
};

export default function ResetPasswordPage(): JSX.Element {
  const { enqueueSnackbar } = useSnackbar();
  const location = useLocation();
  const navigate = useNavigate();
  const resetPassword = useResetPassword();

  const { state: routeState } = useLocation();

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
  }, [routeState, token, enqueueSnackbar, navigate, location.pathname]);

  function validateInput(values: ResetPasswordValues): FormikErrors<ResetPasswordValues> {
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

  async function submitResetPassword(values: ResetPasswordValues, helpers: FormikHelpers<ResetPasswordValues>): Promise<void> {
    helpers.setSubmitting(true);

    return resetPassword(values.password, token)
      // If the reset password fails, then set submitting to false and do nothing. The error will already have been
      // displayed by the reset password function. We only do this if there is an error because if this succeeds then
      // the user is automatically redirected to the login page.
      .catch(() => helpers.setSubmitting(false));
  }

  const message = (routeState && routeState['message']) || 'Enter the new password you would like to use.';

  return (
    <Fragment>
      <BackToLoginButton />
      <Formik
        initialValues={ initialValues }
        validate={ validateInput }
        onSubmit={ submitResetPassword }
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
          <form onSubmit={ handleSubmit } className="h-full overflow-y-auto pb-20">
            <div className="flex items-center justify-center w-full h-full max-h-full">
              <div className="w-full p-10 xl:w-3/12 lg:w-5/12 md:w-2/3 sm:w-10/12 max-w-screen-sm sm:p-0">
                <CenteredLogo />
                <div className="w-full">
                  <div className="w-full pb-2.5">
                    <p className="text-center">
                      { message }
                    </p>
                  </div>
                  <div className="w-full pb-2.5">
                    <TextField
                      autoComplete="current-password"
                      autoFocus
                      className="w-full"
                      disabled={ isSubmitting }
                      error={ touched.password && !!errors.password }
                      helperText={ (touched.password && errors.password) ? errors.password : null }
                      id="reset-password"
                      label="Password"
                      name="password"
                      onBlur={ handleBlur }
                      onChange={ handleChange }
                      value={ values.password }
                      variant="outlined"
                      type="password"
                    />
                  </div>
                  <div className="w-full pb-2.5">
                    <TextField
                      autoComplete="current-password"
                      className="w-full"
                      disabled={ isSubmitting }
                      error={ touched.verifyPassword && !!errors.verifyPassword }
                      helperText={ (touched.verifyPassword && errors.verifyPassword) ? errors.verifyPassword : null }
                      id="reset-password-verify"
                      label="Verify Password"
                      name="verifyPassword"
                      onBlur={ handleBlur }
                      onChange={ handleChange }
                      value={ values.verifyPassword }
                      variant="outlined"
                      type="password"
                    />
                  </div>
                  <div className="w-full pt-2.5 mb-10">
                    <Button
                      className="w-full"
                      color="primary"
                      disabled={ isSubmitting || !values.password || !values.verifyPassword }
                      onClick={ submitForm }
                      type="submit"
                      variant="contained"
                    >
                      { isSubmitting && <CircularProgress
                        className={ classnames('mr-2', {
                          'opacity-50': isSubmitting,
                        }) }
                        size="1em"
                        thickness={ 5 }
                      /> }
                      { isSubmitting ? 'Resetting Password...' : 'Reset Password' }
                    </Button>
                  </div>
                </div>
              </div>
            </div>
          </form>
        ) }
      </Formik>
    </Fragment>
  );
}
