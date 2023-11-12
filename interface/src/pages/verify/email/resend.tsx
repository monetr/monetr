import React, { useState } from 'react';
import { useLocation } from 'react-router-dom';
import { FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import MFormButton from '@monetr/interface/components/MButton';
import MCaptcha from '@monetr/interface/components/MCaptcha';
import MForm from '@monetr/interface/components/MForm';
import MLink from '@monetr/interface/components/MLink';
import MLogo from '@monetr/interface/components/MLogo';
import MSpan from '@monetr/interface/components/MSpan';
import MTextField from '@monetr/interface/components/MTextField';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import request from '@monetr/interface/util/request';
import verifyEmailAddress from '@monetr/interface/util/verifyEmailAddress';

interface ResendValues {
  email: string | null;
  captcha: string | null;
}

export default function ResendVerificationPage(): JSX.Element {
  const { enqueueSnackbar } = useSnackbar();
  const config = useAppConfiguration();
  const { state: routeState } = useLocation();
  const initialValues: ResendValues = {
    email: (routeState && routeState['emailAddress']) || undefined,
    captcha: null,
  };

  const [done, setDone] = useState(false);

  async function resendVerification(values: ResendValues): Promise<void> {
    return request().post('/authentication/verify/resend', {
      email: values.email,
      captcha: values.captcha,
    })
      .then(() => setDone(true))
      .catch(error => void enqueueSnackbar(error?.response?.data?.error || 'Failed to resend verification link', {
        variant: 'error',
        disableWindowBlurListener: true,
      }));
  }

  function validateInput(values: ResendValues): Partial<ResendValues> | null {
    const errors: Partial<ResendValues> = {};

    if (values.email) {
      if (!verifyEmailAddress(values.email)) {
        errors['email'] = 'Please provide a valid email address.';
      }
    }

    return errors;
  }

  async function submit(values: ResendValues, helpers: FormikHelpers<ResendValues>): Promise<void> {
    helpers.setSubmitting(true);
    return resendVerification(values)
      .finally(() => helpers.setSubmitting(false));
  }

  if (done) {
    return <AfterEmailVerificationSent />;
  }

  function RouteStateMessage(): JSX.Element {
    if (routeState) {
      return (
        <MSpan className='text-center' size='sm' data-testid='resend-email-included'>
          It looks like your email address has not been verified. Do you want to resend the email verification link?
        </MSpan>
      );
    }

    return (
      <MSpan className='text-center' size='sm' data-testid='resend-email-excluded'>
        If your email verification link has expired, or you never got one. You can enter your email address below and
        another verification link will be sent to you.
      </MSpan>
    );
  }

  return (
    <MForm
      initialValues={ initialValues }
      onSubmit={ submit }
      validate={ validateInput }
      className='w-full h-full flex flex-col justify-center items-center gap-2 p-4'
    >
      <div className='max-w-xs flex flex-col items-center gap-2'>
        <MLogo className='h-24 w-24' />
        <RouteStateMessage />
        <MTextField
          name='email'
          autoComplete='username'
          autoFocus
          label='Email'
          className='w-full'
          data-testid='resend-email'
        />
        <MCaptcha
          name='captcha'
          // Show the captcha if there is a captcha key specified in the config.
          show={ Boolean(config?.ReCAPTCHAKey) }
          data-testid='resend-captcha'
        />
        <MFormButton type='submit' color='primary' className='w-full'>
          Resend Verification
        </MFormButton>
        <div className='mt-1 flex justify-center gap-1'>
          <MSpan color='subtle' className='text-sm'>Don't need to resend?</MSpan>
          <MLink to='/login' size='sm' data-testid='login-signup'>Return to login</MLink>
        </div>
      </div>
    </MForm>
  );
};

export function AfterEmailVerificationSent(): JSX.Element {
  return (
    <div className='h-full w-full flex flex-col items-center justify-center'>
      <div className='flex flex-col gap-2 max-w-xs items-center'>
        <MLogo className='h-24 w-24' />
        <MSpan className='text-center' size='lg'>
          A new verification link was sent to your email address...
        </MSpan>
        <div className='mt-1 flex justify-center gap-1'>
          <MLink to='/login' size='sm' data-testid='login-signup'>Return to login</MLink>
        </div>
      </div>
    </div>
  );
}
