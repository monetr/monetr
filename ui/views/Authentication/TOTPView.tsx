import React, { Fragment, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { Button, TextField } from '@mui/material';
import { Formik, FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import BackToLoginButton from 'components/Authentication/BackToLoginButton';
import CaptchaMaybe from 'components/Captcha/CaptchaMaybe';
import CircularProgress from 'components/CircularProgress';
import CenteredLogo from 'components/Logo/CenteredLogo';

interface TOTPViewParameters {
  emailAddress: string;
  password: string;
}

interface TOTPFormValues {
  code: string;
}

export default function TOTPView(): JSX.Element {
  const { state: routeState } = useLocation();
  const navigate = useNavigate();
  const { enqueueSnackbar } = useSnackbar();
  const [verification, setVerification] = useState<string | null>();

  // If the user tries to navigate here without these parameters being provided by being routed from somewhere else then
  // kick the user back to the login page.
  if (!routeState['emailAddress'] || !routeState['password']) {
    navigate('/login');
    enqueueSnackbar('Must provide login details to authenticate with MFA.', {
      variant: 'warning',
      disableWindowBlurListener: true,
    });
    return null;
  }

  const input: TOTPViewParameters = {
    emailAddress: routeState['emailAddress'],
    password: routeState['password'],
  };

  const initialValues: TOTPFormValues = { code: '' };

  function submit(values: TOTPFormValues, helpers: FormikHelpers<TOTPFormValues>): Promise<void> {
    helpers.setSubmitting(true);
    return Promise.resolve();
  }

  return (
    <Fragment>
      <BackToLoginButton />
      <Formik
        initialValues={ initialValues }
        onSubmit={ submit }
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
          <form onSubmit={ handleSubmit } className="h-full overflow-y-auto">
            <div className="flex items-center justify-center w-full h-full max-h-full">
              <div className="w-full p-10 xl:w-3/12 lg:w-5/12 md:w-2/3 sm:w-10/12 max-w-screen-sm sm:p-0">
                <CenteredLogo />
                <div className="w-full">
                  <div className="w-full pb-2.5">
                    <p className="text-center">
                      MFA is required to login.
                    </p>
                  </div>
                  <div className="w-full pb-2.5">
                    <TextField
                      autoComplete="totp"
                      autoFocus
                      className="w-full"
                      disabled={ isSubmitting }
                      error={ touched.code && !!errors.code }
                      helperText={ (touched.code && errors.code) ? errors.code : null }
                      id="login-totp"
                      label="Code"
                      name="totp"
                      onBlur={ handleBlur }
                      onChange={ handleChange }
                      value={ values.code }
                      variant="outlined"
                    />
                  </div>
                </div>
                <CaptchaMaybe
                  show
                  loading={ isSubmitting }
                  onVerify={ setVerification }
                />
                <div className="w-full pt-2.5 mb-10">
                  <Button
                    className="w-full"
                    color="primary"
                    disabled={ isSubmitting || !values.code }
                    onClick={ submitForm }
                    type="submit"
                    variant="contained"
                  >
                    <CircularProgress
                      className="mr-2"
                      visible={ isSubmitting }
                      submitting={ isSubmitting }
                      size="1em"
                      thickness={ 5 }
                    />
                    { isSubmitting ? 'Submitting...' : 'Submit' }
                  </Button>
                </div>
              </div>
            </div>
          </form>
        ) }
      </Formik>
    </Fragment>
  );
}
