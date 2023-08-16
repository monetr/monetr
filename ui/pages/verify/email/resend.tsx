import React, { useState } from 'react';
import { useLocation } from 'react-router-dom';
import { FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import AfterEmailVerificationSent from 'components/Authentication/AfterEmailVerificationSent';
import MFormButton from 'components/MButton';
import MCaptcha from 'components/MCaptcha';
import MForm from 'components/MForm';
import MLink from 'components/MLink';
import MLogo from 'components/MLogo';
import MSpan from 'components/MSpan';
import MTextField from 'components/MTextField';
import { useAppConfiguration } from 'hooks/useAppConfiguration';
import request from 'util/request';
import verifyEmailAddress from 'util/verifyEmailAddress';

interface ResendValues {
  email: string | null;
  captcha: string | null;
}

export default function ResendVerificationPage(): JSX.Element {
  const { enqueueSnackbar } = useSnackbar();
  const config = useAppConfiguration();
  const { state: routeState } = useLocation();
  const initialValues: ResendValues = {
    email: (routeState && routeState['emailAddress']) || null,
    captcha: null,
  };

  const [done, setDone] = useState(false);

  async function resendVerification(values: ResendValues): Promise<void> {
    return request().post('/authentication/verify/resend', {
      email: values.captcha,
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
        <MSpan className='text-center' size='sm'>
          It looks like your email address has not been verified. Do you want to resend the email verification link?
        </MSpan>
      );
    }

    return (
      <MSpan className='text-center' size='sm'>
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
          name="email"
          autoComplete="username"
          autoFocus
          label="Email"
          className='w-full'
        />
        <MCaptcha
          name="captcha"
          // Show the captcha if there is a captcha key specified in the config.
          show={ Boolean(config?.ReCAPTCHAKey) }
        />
        <MFormButton type='submit' color='primary' className='w-full'>
          Resend Verification
        </MFormButton>
        <div className="mt-1 flex justify-center gap-1">
          <MSpan variant="light" className='text-sm'>Don't need to resend?</MSpan>
          <MLink to="/login" size="sm" data-testid='login-signup'>Return to login</MLink>
        </div>
      </div>
    </MForm>
  );
};
