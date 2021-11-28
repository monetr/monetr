import TextWithLine from 'components/TextWithLine';
import { useSnackbar } from 'notistack';
import React, { Fragment, useState } from 'react';
import { useSelector } from 'react-redux';
import { Link as RouterLink } from 'react-router-dom';
import useLogin from 'shared/authentication/actions/login';
import { getReCAPTCHAKey, getShouldVerifyLogin, getSignUpAllowed } from 'shared/bootstrap/selectors';
import classnames from 'classnames';
import {
  Button,
  CircularProgress,
  TextField
} from '@mui/material';
import { Formik, FormikHelpers } from 'formik';
import verifyEmailAddress from 'util/verifyEmailAddress';
import CaptchaMaybe from 'views/Captcha/CaptchaMaybe';

import Logo from 'assets';

interface LoginValues {
  email: string | null;
  password: string | null;
}

const LoginView = (): JSX.Element => {
  const { enqueueSnackbar } = useSnackbar();
  const ReCAPTCHAKey = useSelector(getReCAPTCHAKey);
  const allowSignUp = useSelector(getSignUpAllowed);
  const verifyLogin = useSelector(getShouldVerifyLogin);

  const login = useLogin();

  const [captcha, setCaptcha] = useState<string | null>(null);

  function validateInput(values: LoginValues): Partial<LoginValues> {
    let errors: Partial<LoginValues> = {};

    if (values.email) {
      if (!verifyEmailAddress(values.email)) {
        errors['email'] = 'Please provide a valid email address.';
      }
    }

    if (values.password) {
      if (values.password.length < 8) {
        errors['password'] = 'Password must be at least 8 characters long.'
      }
    }

    return errors;
  }

  function doLogin(values: LoginValues, helpers: FormikHelpers<LoginValues>) {
    helpers.setSubmitting(true);

    return login({
      captcha: captcha,
      email: values.email,
      password: values.password
    })
      .catch(error => enqueueSnackbar(error?.response?.data?.error || 'Failed to authenticate.', {
        variant: 'error',
        disableWindowBlurListener: true,
      }))
      .finally(() => helpers.setSubmitting(false));
  }

  function renderBottomButtons(
    isSubmitting: boolean,
    disableForVerification: boolean,
    values: LoginValues,
    submitForm: () => Promise<any>,
  ): JSX.Element {
    return (
      <div>
        <div className="w-full pt-2.5 pb-2.5">
          <Button
            className="w-full"
            color="primary"
            disabled={ isSubmitting || (!values.password || !values.email || !disableForVerification) }
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
            { isSubmitting ? 'Signing In...' : 'Sign In' }
          </Button>
        </div>
      </div>
    )
  }

  const initialValues: LoginValues = {
    email: '',
    password: '',
  }

  const disableForVerification = !verifyLogin || Boolean(ReCAPTCHAKey && captcha);

  return (
    <Fragment>
      <Formik
        initialValues={ initialValues }
        validate={ validateInput }
        onSubmit={ doLogin }
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
                <div className="flex justify-center w-full mb-5">
                  <img src={ Logo } className="w-1/3"/>
                </div>
                { allowSignUp && (
                  <div>
                    <div className="w-full pb-2.5">
                      <Button
                        className="w-full"
                        color="secondary"
                        component={ RouterLink }
                        disabled={ isSubmitting }
                        to="/register"
                        variant="contained"
                      >
                        Sign Up For monetr
                      </Button>
                    </div>
                    <div className="w-full opacity-50 pb-2.5">
                      <TextWithLine>
                        or sign in with your email
                      </TextWithLine>
                    </div>
                  </div>
                ) }
                <div className="w-full">
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
                  <div className="w-full pt-2.5 pb-2.5">
                    <TextField
                      autoComplete="current-password"
                      className="w-full"
                      disabled={ isSubmitting }
                      error={ touched.password && !!errors.password }
                      helperText={ (touched.password && errors.password) ? errors.password : null }
                      id="login-password"
                      label="Password"
                      name="password"
                      onBlur={ handleBlur }
                      onChange={ handleChange }
                      type="password"
                      value={ values.password }
                      variant="outlined"
                    />
                  </div>
                </div>
                <CaptchaMaybe
                  loading={ isSubmitting }
                  show={ verifyLogin }
                  onVerify={ setCaptcha }
                />
                { renderBottomButtons(isSubmitting, disableForVerification, values, submitForm) }
              </div>
            </div>
          </form>
        ) }
      </Formik>
    </Fragment>
  )
};

export default LoginView;
