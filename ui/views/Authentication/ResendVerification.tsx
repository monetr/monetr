import { useSnackbar } from 'notistack';
import React, { Fragment, useState } from 'react';
import { useSelector } from 'react-redux';
import { Formik, FormikHelpers } from 'formik';
import { Button, CircularProgress, TextField } from '@mui/material';
import classnames from 'classnames';
import { useLocation } from 'react-router-dom';
import AfterEmailVerificationSent from 'views/Authentication/AfterEmailVerificationSent';
import AuthenticationLogo from 'views/Authentication/components/AuthenticationLogo';
import BackToLoginButton from 'views/Authentication/components/BackToLoginButton';
import CaptchaMaybe from 'views/Captcha/CaptchaMaybe';
import verifyEmailAddress from 'util/verifyEmailAddress';
import request from 'shared/util/request';
import { getReCAPTCHAKey } from 'shared/bootstrap/selectors';

interface ResendValues {
  email: string | null;
}

const ResendVerification = (): JSX.Element => {
  const { enqueueSnackbar } = useSnackbar();
  const requireCaptcha = !!useSelector(getReCAPTCHAKey);
  const { state: routeState } = useLocation();
  const initialValues: ResendValues = {
    email: (routeState && routeState['emailAddress']) || null,
  }

  const [verification, setVerification] = useState<string | null>();
  const [done, setDone] = useState(false);

  function resendVerification(emailAddress: string): Promise<void> {
    return request().post('/authentication/verify/resend', {
      email: emailAddress,
      captcha: verification,
    })
      .then(() => setDone(true))
      .catch(error => void enqueueSnackbar(error?.response?.data?.error || 'Failed to resend verification link', {
        variant: 'error',
        disableWindowBlurListener: true,
      }))
  }

  function validateInput(values: ResendValues): Partial<ResendValues> | null {
    let errors: Partial<ResendValues> = {};

    if (values.email) {
      if (!verifyEmailAddress(values.email)) {
        errors['email'] = 'Please provide a valid email address.';
      }
    }

    return errors;
  }

  function submit(values: ResendValues, helpers: FormikHelpers<ResendValues>): Promise<void> {
    helpers.setSubmitting(true);
    return resendVerification(values.email)
      .finally(() => helpers.setSubmitting(false));
  }

  if (done) {
    return <AfterEmailVerificationSent/>;
  }

  return (
    <Fragment>
      <BackToLoginButton/>
      <Formik
        initialValues={ initialValues }
        validate={ validateInput }
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
                <AuthenticationLogo/>
                <div className="w-full">
                  <div className="w-full pb-2.5">
                    { routeState &&
                    <p className="text-center">
                      It looks like your email address has not been verified. Do you want to resend the email
                      verification link?
                    </p>
                    }

                    { !routeState &&
                    <p className="text-center">
                      If your email verification link has expired, or you never got one. You can enter your email
                      address below and another verification link will be sent to you.
                    </p>
                    }
                  </div>
                  <div className="w-full pb-2.5">
                    <TextField
                      autoComplete="username"
                      autoFocus
                      className="w-full"
                      disabled={ isSubmitting }
                      error={ touched.email && !!errors.email }
                      helperText={ (touched.email && errors.email) ? errors.email : null }
                      id="login-email"
                      label="Email"
                      name="email"
                      onBlur={ handleBlur }
                      onChange={ handleChange }
                      value={ values.email }
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
                    disabled={ isSubmitting || !values.email || (requireCaptcha && !verification) }
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
                    { isSubmitting ? 'Sending Verification Link...' : 'Resend Verification Link' }
                  </Button>
                </div>
              </div>
            </div>
          </form>
        ) }
      </Formik>
    </Fragment>
  );
};

export default ResendVerification;
