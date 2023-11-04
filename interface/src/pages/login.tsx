import React from 'react';
import { FormikErrors, FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import MFormButton from '@monetr/interface/components/MButton';
import MCaptcha from '@monetr/interface/components/MCaptcha';
import MForm from '@monetr/interface/components/MForm';
import MLink from '@monetr/interface/components/MLink';
import MLogo from '@monetr/interface/components/MLogo';
import MSpan from '@monetr/interface/components/MSpan';
import MTextField from '@monetr/interface/components/MTextField';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import useLogin from '@monetr/interface/hooks/useLogin';
import verifyEmailAddress from '@monetr/interface/util/verifyEmailAddress';

interface LoginValues {
  email: string;
  password: string;
  captcha: string | null;
}

const initialValues: LoginValues = {
  email: '',
  password: '',
  captcha: null,
};

function validator(values: LoginValues): FormikErrors<LoginValues> {
  const errors = {};

  if (values?.email.length === 0) {
    errors['email'] = 'Email must be provided.';
  }

  if (values?.email && !verifyEmailAddress(values?.email)) {
    errors['email'] = 'Email must be valid.';
  }

  if (values?.password.length < 8) {
    errors['password'] = 'Password must be at least 8 characters long.';
  }

  return errors;
}

export default function Login(): JSX.Element {
  const { enqueueSnackbar } = useSnackbar();
  const config = useAppConfiguration();
  const login = useLogin();

  function ForgotPasswordButton(): JSX.Element {
    // If the application is not configured to allow forgot password then don't show the button.
    if (!config?.allowForgotPassword) {
      return null;
    }

    return (
      <MLink to="/password/forgot" size="sm" data-testid='login-forgot' tabIndex={ 5 }>
        Forgot password?
      </MLink>
    );
  }

  function SignUpButton(): JSX.Element {
    if (!config?.allowSignUp) return null;

    return (
      <div className="w-full lg:w-1/4 sm:w-1/3 mt-1 flex justify-center gap-1">
        <MSpan color="subtle" className='text-sm'>Not a user?</MSpan>
        <MLink to="/register" size="sm" data-testid='login-signup' tabIndex={ 4 }>Sign up now</MLink>
      </div>
    );
  }

  async function submit(values: LoginValues, helpers: FormikHelpers<LoginValues>) {
    helpers.setSubmitting(true);

    return login({
      captcha: values.captcha,
      email: values.email,
      password: values.password,
    })
      .catch(error => enqueueSnackbar(error?.response?.data?.error || 'Failed to authenticate.', {
        variant: 'error',
        disableWindowBlurListener: true,
      }))
      .finally(() => helpers.setSubmitting(false));
  }

  return (
    <MForm
      initialValues={ initialValues }
      validate={ validator }
      onSubmit={ submit }
      className="w-full h-full flex pt-10 md:pt-0 md:pb-10 md:justify-center items-center flex-col gap-1 px-5"
    >
      <div className="max-w-[128px] w-full">
        <MLogo />
      </div>
      <MSpan>Sign into your monetr account</MSpan>
      <MTextField
        data-testid='login-email'
        autoFocus
        label="Email Address"
        name='email'
        type='email'
        required
        className="w-full xl:w-1/5 lg:w-1/4 md:w-1/3 sm:w-1/2"
        tabIndex={ 1 }
      />
      <MTextField
        autoComplete='current-password'
        className="w-full xl:w-1/5 lg:w-1/4 md:w-1/3 sm:w-1/2"
        data-testid='login-password'
        label="Password"
        labelDecorator={ ForgotPasswordButton }
        name='password'
        required
        type='password'
        tabIndex={ 2 }
      />
      <MCaptcha
        name="captcha"
        show={ Boolean(config?.verifyLogin) }
      />
      <div className="w-full xl:w-1/5 lg:w-1/4 md:w-1/3 sm:w-1/2 mt-1">
        <MFormButton
          data-testid='login-submit'
          color="primary"
          variant="solid"
          role="form"
          type="submit"
          className='w-full'
          tabIndex={ 3 }
        >
          Sign In
        </MFormButton>
      </div>
      <SignUpButton />
    </MForm>
  );
}

