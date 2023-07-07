import React, { useState } from 'react';
import { useQueryClient } from 'react-query';
import { useNavigate } from 'react-router-dom';
import { AxiosError } from 'axios';
import { Formik, FormikErrors, FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import MFormButton from 'components/MButton';
import MCaptcha from 'components/MCaptcha';
import MCheckbox from 'components/MCheckbox';
import MForm from 'components/MForm';
import MLink from 'components/MLink';
import MLogo from 'components/MLogo';
import MSpan from 'components/MSpan';
import MTextField from 'components/MTextField';
import { useAppConfiguration } from 'hooks/useAppConfiguration';
import useSignUp, { SignUpResponse } from 'hooks/useSignUp';
import { APIError } from 'util/request';
import verifyEmailAddress from 'util/verifyEmailAddress';

interface RegisterValues {
  agree: boolean;
  firstName: string;
  lastName: string;
  email: string;
  password: string;
  confirmPassword: string;
  captcha?: string;
  betaCode?: string;
}

const initialValues: RegisterValues = {
  agree: false,
  firstName: '',
  lastName: '',
  email: '',
  password: '',
  confirmPassword: '',
};

function validator(values: RegisterValues): FormikErrors<RegisterValues> {
  const errors = {};

  if (values?.firstName.length < 2) {
    errors['firstName'] = 'First name must have at least 2 characters.';
  }

  if (values?.lastName.length < 2) {
    errors['lastName'] = 'Last name must have at least 2 characters.';
  }

  if (values?.email.length === 0) {
    errors['email'] = 'Email must be provided.';
  }

  if (values?.email && !verifyEmailAddress(values?.email)) {
    errors['email'] = 'Email must be valid.';
  }

  if (values?.password.length < 8) {
    errors['password'] = 'Password must be at least 8 characters long.';
  }

  if (values?.confirmPassword !== values?.password) {
    errors['confirmPassword'] = 'Password confirmation must match.';
  }

  return errors;
}

export function RegisterSuccessful(): JSX.Element {
  // TODO Add a link to return to the login page, or close this window maybe?
  return (
    <div className='w-full h-full flex justify-center items-center flex-col'>
      <MLogo className='h-24 w-24' />
      <MSpan className='text-xl font-medium max-w-md text-center'>
        A verification message has been sent to your email address, please verify your email.
      </MSpan>
    </div>
  );
}

export default function RegisterNew(): JSX.Element {
  const { enqueueSnackbar } = useSnackbar();
  const config = useAppConfiguration();
  const signUp = useSignUp();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [successful, setSuccessful] = useState(false);

  function BetaCodeInput(): JSX.Element {
    if (!config?.requireBetaCode) return null;

    return (
      <MTextField
        label="Beta Code"
        name="betaCode"
        type="text"
        required
        uppercasetext
        className="xl:w-2/5 lg:w-1/3 md:w-1/2 w-full"
      />
    );
  }

  async function submit(values: RegisterValues, helpers: FormikHelpers<RegisterValues>): Promise<void> {
    helpers.setSubmitting(true);
    return signUp({
      agree: values.agree,
      betaCode: values.betaCode,
      captcha: values.captcha,
      email: values.email,
      firstName: values.firstName,
      lastName: values.lastName,
      password: values.password,
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    })
      .then((result: SignUpResponse) => {
        // After sending the sign up request, if the user needs to verify their email then the requires verification
        // field will be true. We can stop here and just show the user a successful screen.
        if (result.requireVerification) {
          return setSuccessful(true);
        }

        return queryClient.invalidateQueries('/users/me')
          .then(() => {
            if (result.nextUrl) {
              return navigate(result.nextUrl);
            }

            return navigate('/');
          });
      })
      .catch((error: AxiosError<APIError>) => {
        const message = error.response.data.error || 'Failed to sign up.';
        enqueueSnackbar(message, {
          variant: 'error',
          disableWindowBlurListener: true,
        });
      })
      .finally(() => helpers.setSubmitting(false));
  }

  if (successful) {
    return (
      <RegisterSuccessful />
    );
  }

  return (
    <Formik
      initialValues={ initialValues }
      validate={ validator }
      onSubmit={ submit }
    >
      <MForm className="w-full h-full flex pt-10 md:pt-0 md:pb-10 md:justify-center items-center flex-col gap-1 px-5">
        <div className="max-w-[96px] w-full">
          <MLogo />
        </div>
        <div className="flex flex-col items-center">
          <MSpan className='text-5xl'>
            Get Started
          </MSpan>
          <MSpan variant="light" className='text-lg'>
            Create your monetr account now
          </MSpan>
        </div>
        <div className="flex flex-col sm:flex-row gap-2.5 xl:w-2/5 lg:w-1/3 md:w-1/2 w-full">
          <MTextField
            autoFocus
            label="First Name"
            name="firstName"
            type="text"
            required
            className="w-full"
          />
          <MTextField
            label="Last Name"
            name="lastName"
            type="text"
            required
            className="w-full"
          />
        </div>
        <MTextField
          label="Email Address"
          name='email'
          type='email'
          required
          className="xl:w-2/5 lg:w-1/3 md:w-1/2 w-full"
        />
        <MTextField
          label="Password"
          name='password'
          type='password'
          required
          className="xl:w-2/5 lg:w-1/3 md:w-1/2 w-full"
        />
        <MTextField
          label="Confirm Password"
          name='confirmPassword'
          type='password'
          required
          className="xl:w-2/5 lg:w-1/3 md:w-1/2 w-full"
        />
        <BetaCodeInput />
        <MCaptcha
          className='mb-1'
          name="captcha"
          show={ Boolean(config?.verifyRegister) }
        />
        <MCheckbox
          id="terms"
          name="agree"
          label={
            <MSpan>
              I agree to monetr's&nbsp;
              <a
                target="_blank"
                className="text-blue-500 hover:underline focus:ring-2 focus:ring-blue-500 focus:underline"
                href='https://github.com/monetr/legal/blob/main/TERMS_OF_USE.md'>
                Terms of Use
              </a> and&nbsp;
              <a
                target="_blank"
                className="text-blue-500 hover:underline focus:ring-2 focus:ring-blue-500 focus:underline"
                href='https://github.com/monetr/legal/blob/main/PRIVACY.md'
              >
                Privacy Policy
              </a>
            </MSpan>
          }
        />
        <div className="xl:w-2/5 lg:w-1/3 md:w-1/2 w-full mt-1">
          <MFormButton color="primary" variant="solid" role="form" type="submit" className='w-full'>
            Sign Up
          </MFormButton>
        </div>
        <div className="xl:w-2/5 lg:w-1/3 md:w-1/2 w-full mt-1 flex justify-center gap-1">
          <MSpan variant="light" className='text-sm'>Already have an account?</MSpan>
          <MLink to="/login" size="sm">Sign in instead</MLink>
        </div>
      </MForm>
    </Formik>
  );
}
