import { useCallback, useState } from 'react';
import type { FormikErrors, FormikHelpers } from 'formik';
import { useSearch } from 'wouter';

import type { ApiError } from '@monetr/interface/api/client';
import FormButton from '@monetr/interface/components/FormButton';
import FormTextField from '@monetr/interface/components/FormTextField';
import MForm from '@monetr/interface/components/MForm';
import MLogo from '@monetr/interface/components/MLogo';
import TextLink from '@monetr/interface/components/TextLink';
import Typography from '@monetr/interface/components/Typography';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import { useProofOfWork } from '@monetr/interface/hooks/useProofOfWork';
import request, { type APIError } from '@monetr/interface/util/request';
import verifyEmailAddress from '@monetr/interface/util/verifyEmailAddress';
import { useSnackbar } from '@monetr/notify';

import styles from './resend.module.scss';

interface ResendValues {
  email: string | null;
}

export default function ResendVerificationPage(): React.JSX.Element {
  const { enqueueSnackbar } = useSnackbar();
  const { data: config } = useAppConfiguration();
  const emailFromQuery = new URLSearchParams(useSearch()).get('email');
  const [done, setDone] = useState(false);
  const pow = useProofOfWork('resend', Boolean(config?.proofOfWorkEnabled));

  const resendVerification = useCallback(
    async (values: ResendValues): Promise<void> => {
      // getSolution is null when proof of work is disabled (challenge/nonce drop
      // off). Kept in the chain so a fetch/solve failure also hits the catch.
      return pow
        .getSolution()
        .then(solution =>
          request({
            method: 'POST',
            url: '/api/authentication/verify/resend',
            data: {
              email: values.email,
              challenge: solution?.challenge,
              nonce: solution?.nonce,
            },
          }),
        )
        .then(() => setDone(true))
        .catch((error: ApiError<APIError>) => {
          // Single use and consumed even on failure, so line up a fresh one.
          pow.reset();
          enqueueSnackbar(error?.response?.data?.error || 'Failed to resend verification link', {
            variant: 'error',
            disableWindowBlurListener: true,
          });
        });
    },
    [enqueueSnackbar, pow],
  );

  const validateInput = useCallback((values: ResendValues): FormikErrors<ResendValues> => {
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
    email: emailFromQuery || null,
  };

  if (done) {
    return <AfterEmailVerificationSent />;
  }

  return (
    <MForm
      className={styles.form}
      initialValues={initialValues}
      onInput={pow.warmup}
      onSubmit={submit}
      validate={validateInput}
    >
      <div className={styles.panel}>
        <MLogo className={styles.logo} />
        <RouteStateMessage hasEmail={Boolean(emailFromQuery)} />
        <FormTextField
          autoComplete='username'
          autoFocus
          className={styles.field}
          data-testid='resend-email'
          label='Email'
          name='email'
        />
        <FormButton className={styles.field} color='primary' type='submit'>
          Resend Verification
        </FormButton>
        <div className={styles.loginRow}>
          <Typography color='subtle' size='sm'>
            Don&apos;t need to resend?
          </Typography>
          <TextLink data-testid='login-signup' size='sm' to='/login'>
            Return to login
          </TextLink>
        </div>
      </div>
    </MForm>
  );
}

export function AfterEmailVerificationSent(): React.JSX.Element {
  return (
    <div className={styles.sentRoot}>
      <div className={styles.panel}>
        <MLogo className={styles.logo} />
        <Typography align='center' size='lg'>
          A new verification link was sent to your email address...
        </Typography>
        <div className={styles.loginRow}>
          <TextLink data-testid='login-signup' size='sm' to='/login'>
            Return to login
          </TextLink>
        </div>
      </div>
    </div>
  );
}

interface RouteStateMessageProps {
  hasEmail: boolean;
}

function RouteStateMessage(props: RouteStateMessageProps): React.JSX.Element {
  if (props.hasEmail) {
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
