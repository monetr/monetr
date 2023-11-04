/* eslint-disable max-len */
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';
import { AxiosError } from 'axios';
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
import useSignUp, { SignUpResponse } from '@monetr/interface/hooks/useSignUp';
import { APIError } from '@monetr/interface/util/request';
import verifyEmailAddress from '@monetr/interface/util/verifyEmailAddress';

interface RegisterValues {
  firstName: string;
  lastName: string;
  email: string;
  password: string;
  confirmPassword: string;
  captcha?: string;
  betaCode?: string;
}

const initialValues: RegisterValues = {
  firstName: '',
  lastName: '',
  email: '',
  password: '',
  confirmPassword: '',
};

const breakpoints = 'w-full md:w-1/2 lg:w-1/3 xl:w-1/4';

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

  // TODO No restriction on agree?
  return errors;
}

export function RegisterSuccessful(): JSX.Element {
  // TODO Add a link to return to the login page, or close this window maybe?
  return (
    <div className='w-full h-full flex justify-center items-center flex-col'>
      <MLogo className='h-24 w-24' />
      <MSpan size='xl' weight='medium' className='max-w-md text-center'>
        A verification message has been sent to your email address, please verify your email.
      </MSpan>
    </div>
  );
}

export default function Register(): JSX.Element {
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
        className={ breakpoints }
      />
    );
  }

  async function submit(values: RegisterValues, helpers: FormikHelpers<RegisterValues>): Promise<void> {
    helpers.setSubmitting(true);

    return signUp({
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

        return queryClient.invalidateQueries(['/users/me'])
          .then(() => {
            // If the register endpoint has told us to navigate to a specific url afterwards, then do that now.
            if (result.nextUrl) {
              return navigate(result.nextUrl);
            }

            // Otherwise just go to the index-ish route for the authenticated app.
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
    <div className='w-full h-full flex pt-10 md:pt-0 md:pb-10 md:justify-center items-center flex-col gap-1 px-5 overflow-y-auto py-4'>
      <MForm
        initialValues={ initialValues }
        validate={ validator }
        onSubmit={ submit }
        className="flex flex-col md:w-1/2 lg:w-1/3 xl:w-1/4 items-center"
      >
        <div className="max-w-[96px] w-full">
          <MLogo />
        </div>
        <div className="flex flex-col items-center text-center">
          <MSpan className='text-5xl'>
            Get Started
          </MSpan>
          <MSpan color="subtle" className='text-lg'>
            Create your monetr account now
          </MSpan>
        </div>
        <div className="flex flex-col sm:flex-row gap-2.5 w-full">
          <MTextField
            data-testid='register-first-name'
            autoFocus
            label="First Name"
            name="firstName"
            type="text"
            required
            className="w-full"
          />
          <MTextField
            data-testid='register-last-name'
            label="Last Name"
            name="lastName"
            type="text"
            required
            className="w-full"
          />
        </div>
        <MTextField
          data-testid='register-email'
          label="Email Address"
          name='email'
          type='email'
          required
          className="w-full"
        />
        <MTextField
          autoComplete='new-password'
          className="w-full"
          data-testid='register-password'
          label="Password"
          name='password'
          required
          type='password'
        />
        <MTextField
          autoComplete='new-password'
          className="w-full"
          data-testid='register-confirm-password'
          label="Confirm Password"
          name='confirmPassword'
          required
          type='password'
        />
        <BetaCodeInput />
        <MCaptcha
          className='mb-1'
          name="captcha"
          show={ Boolean(config?.verifyRegister) }
        />
        <MFormButton
          data-testid='register-submit'
          className='w-full mt-1'
          color="primary"
          role="form"
          type="submit"
          variant="solid"
        >
          Sign Up
        </MFormButton>
        <div className="mt-1 flex justify-center gap-1 flex-col md:flex-row items-center">
          <MSpan className='gap-1 inline-block text-center' size='sm' color='subtle' component='p'>
            By signing up you agree to monetr's&nbsp;
            <a
              target="_blank"
              className="text-dark-monetr-blue hover:underline focus:ring-2 focus:ring-dark-monetr-blue focus:underline"
              href='https://github.com/monetr/legal/blob/main/TERMS_OF_USE.md'>
              Terms of Use
            </a> and&nbsp;
            <a
              target="_blank"
              className="text-dark-monetr-blue hover:underline focus:ring-2 focus:ring-dark-monetr-blue focus:underline"
              href='https://github.com/monetr/legal/blob/main/PRIVACY.md'
            >
              Privacy Policy
            </a>
          </MSpan>
        </div>
        <div className="mt-1 flex justify-center gap-1 flex-col md:flex-row items-center">
          <MSpan color="subtle" className='text-sm'>Already have an account?</MSpan>
          <MLink to="/login" size="sm">Sign in instead</MLink>
        </div>
      </MForm>
    </div>
  );
}
