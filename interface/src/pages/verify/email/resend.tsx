import { useCallback, useState } from 'react';
import type { AxiosError } from 'axios';
import type { FormikErrors, FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';
import { useLocation } from 'react-router-dom';

import FormButton from '@monetr/interface/components/FormButton';
import FormTextField from '@monetr/interface/components/FormTextField';
import MCaptcha from '@monetr/interface/components/MCaptcha';
import MForm from '@monetr/interface/components/MForm';
import MLogo from '@monetr/interface/components/MLogo';
import TextLink from '@monetr/interface/components/TextLink';
import Typography from '@monetr/interface/components/Typography';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import request, { type APIError } from '@monetr/interface/util/request';
import verifyEmailAddress from '@monetr/interface/util/verifyEmailAddress';

interface ResendValues {
  email: string | null;
  captcha: string | null;
}

export default function ResendVerificationPage(): JSX.Element {
  const { enqueueSnackbar } = useSnackbar();
  const { data: config } = useAppConfiguration();
  const { state: routeState } = useLocation();
  const [done, setDone] = useState(false);

  const resendVerification = useCallback(
    async (values: ResendValues): Promise<void> => {
      return await request()
        .post('/authentication/verify/resend', {
          email: values.email,
          captcha: values.captcha,
        })
        .then(() => setDone(true))
        .catch(
          (error: AxiosError<APIError>) =>
            void enqueueSnackbar(error?.response?.data?.error || 'Failed to resend verification link', {
              variant: 'error',
              disableWindowBlurListener: true,
            }),
        );
    },
    [enqueueSnackbar],
  );

  const validateInput = useCallback((values: ResendValues): FormikErrors<ResendValues> | null => {
    const errors: FormikErrors<ResendValues> = {};

    if (values.email) {
      if (!verifyEmailAddress(values.email)) {
        errors.email = 'Please provide a valid email address.';
      }
    }

    return errors;
  }, []);

  const submit = useCallback(
    async (values: ResendValues, helpers: FormikHelpers<ResendValues>): Promise<void> => {
      helpers.setSubmitting(true);
      return await resendVerification(values).finally(() => helpers.setSubmitting(false));
    },
    [resendVerification],
  );

  const initialValues: ResendValues = {
    email: routeState?.emailAddress || undefined,
    captcha: null,
  };

  if (done) {
    return <AfterEmailVerificationSent />;
  }

  return (
    <MForm
      className='w-full h-full flex flex-col justify-center items-center gap-2 p-4'
      initialValues={initialValues}
      onSubmit={submit}
      validate={validateInput}
    >
      <div className='max-w-xs flex flex-col items-center gap-2'>
        <MLogo className='h-24 w-24' />
        <RouteStateMessage />
        <FormTextField
          autoComplete='username'
          autoFocus
          className='w-full'
          data-testid='resend-email'
          label='Email'
          name='email'
        />
        <MCaptcha
          data-testid='resend-captcha'
          // Show the captcha if there is a captcha key specified in the config.
          name='captcha'
          show={Boolean(config?.ReCAPTCHAKey)}
        />
        <FormButton className='w-full' color='primary' type='submit'>
          Resend Verification
        </FormButton>
        <div className='mt-1 flex justify-center gap-1'>
          <Typography color='subtle' size='sm'>
            Don't need to resend?
          </Typography>
          <TextLink data-testid='login-signup' size='sm' to='/login'>
            Return to login
          </TextLink>
        </div>
      </div>
    </MForm>
  );
}

export function AfterEmailVerificationSent(): JSX.Element {
  return (
    <div className='h-full w-full flex flex-col items-center justify-center'>
      <div className='flex flex-col gap-2 max-w-xs items-center'>
        <MLogo className='h-24 w-24' />
        <Typography align='center' size='lg'>
          A new verification link was sent to your email address...
        </Typography>
        <div className='mt-1 flex justify-center gap-1'>
          <TextLink data-testid='login-signup' size='sm' to='/login'>
            Return to login
          </TextLink>
        </div>
      </div>
    </div>
  );
}

function RouteStateMessage(): JSX.Element {
  const { state: routeState } = useLocation();
  if (routeState) {
    return (
      <Typography align='center' data-testid='resend-email-included' size='sm'>
        It looks like your email address has not been verified. Do you want to resend the email verification link?
      </Typography>
    );
  }

  return (
    <Typography align='center' data-testid='resend-email-excluded' size='sm'>
      If your email verification link has expired, or you never got one. You can enter your email address below and
      another verification link will be sent to you.
    </Typography>
  );
}
