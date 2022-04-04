import { Button, CircularProgress, TextField } from '@mui/material';
import classnames from 'classnames';
import CenteredLogo from 'components/Logo/CenteredLogo';
import { Formik, FormikErrors, FormikHelpers } from 'formik';
import React, { Fragment, useState } from 'react';
import { useSelector } from 'react-redux';
import useSendForgotPassword from 'shared/authentication/actions/sendForgotPassword';
import { getShouldVerifyForgotPassword } from 'shared/bootstrap/selectors';
import verifyEmailAddress from 'util/verifyEmailAddress';
import BackToLoginButton from 'components/Authentication/BackToLoginButton';
import CaptchaMaybe from 'components/Captcha/CaptchaMaybe';

interface ForgotPasswordValues {
  email: string;
}

const initialValues: ForgotPasswordValues = {
  email: '',
};

export default function ForgotPasswordPage(): JSX.Element {
  const verifyForgotPassword = useSelector(getShouldVerifyForgotPassword);
  const [verification, setVerification] = useState<string | null>(null);
  const sendForgotPassword = useSendForgotPassword();

  function validateInput(values: ForgotPasswordValues): FormikErrors<ForgotPasswordValues> {
    let errors: FormikErrors<ForgotPasswordValues> = {};

    if (values.email) {
      if (!verifyEmailAddress(values.email)) {
        errors['email'] = 'Please provide a valid email address.';
      }
    }

    return errors;
  }

  function submitForgotPassword(values: ForgotPasswordValues, helpers: FormikHelpers<ForgotPasswordValues>): Promise<void> {
    helpers.setSubmitting(true);

    // sendForgotPassword pretty much does all the work, the only thing we need to do is make sure that once we are done
    // we set submitting back to false.
    return sendForgotPassword(values.email, verification)
      .finally(() => helpers.setSubmitting(false));
  }

  return (
    <Fragment>
      <BackToLoginButton/>
      <Formik
        initialValues={ initialValues }
        validate={ validateInput }
        onSubmit={ submitForgotPassword }
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
                <CenteredLogo/>
                <div className="w-full">
                  <div className="w-full pb-2.5">
                    <p className="text-center">
                      In order to reset your forgotten password, we will send you an email with a link.
                    </p>
                    <p className="text-center">
                      Please enter the email address for your login below.
                    </p>
                  </div>
                  <div className="w-full pb-2.5">
                    <TextField
                      autoComplete="username"
                      autoFocus
                      className="w-full"
                      disabled={ isSubmitting }
                      error={ touched.email && !!errors.email }
                      helperText={ (touched.email && errors.email) ? errors.email : null }
                      id="forgot-email"
                      label="Email"
                      name="email"
                      onBlur={ handleBlur }
                      onChange={ handleChange }
                      value={ values.email }
                      variant="outlined"
                    />
                  </div>
                  <CaptchaMaybe
                    show={ verifyForgotPassword }
                    loading={ isSubmitting }
                    onVerify={ setVerification }
                  />
                  <div className="w-full pt-2.5 mb-10">
                    <Button
                      className="w-full"
                      color="primary"
                      disabled={ isSubmitting || !values.email || (verifyForgotPassword && !verification) }
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
                      { isSubmitting ? 'Sending Password Reset Link...' : 'Send Password Reset Link' }
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
