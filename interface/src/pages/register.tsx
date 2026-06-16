import { useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import type { FormikErrors, FormikHelpers } from 'formik';
import { useLocation } from 'wouter';

import type { ApiError } from '@monetr/interface/api/client';
import Flex from '@monetr/interface/components/Flex';
import FormButton from '@monetr/interface/components/FormButton';
import FormTextField from '@monetr/interface/components/FormTextField';
import { layoutVariants } from '@monetr/interface/components/Layout';
import MCaptcha from '@monetr/interface/components/MCaptcha';
import MForm from '@monetr/interface/components/MForm';
import MLogo from '@monetr/interface/components/MLogo';
import BetaCodeInput from '@monetr/interface/components/register/BetaCodeInput';
import TextLink from '@monetr/interface/components/TextLink';
import Typography from '@monetr/interface/components/Typography';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import { useProofOfWork } from '@monetr/interface/hooks/useProofOfWork';
import useSignUp, { type SignUpResponse } from '@monetr/interface/hooks/useSignUp';
import { getLocale, getTimezone } from '@monetr/interface/util/locale';
import type { APIError } from '@monetr/interface/util/request';
import verifyEmailAddress from '@monetr/interface/util/verifyEmailAddress';
import { useSnackbar } from '@monetr/notify';

import styles from './register.module.scss';

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

function validator(values: RegisterValues): FormikErrors<RegisterValues> {
  const errors: FormikErrors<RegisterValues> = {};

  if (values?.firstName.length < 2) {
    errors.firstName = 'First name must have at least 2 characters.';
  }

  if (values?.lastName.length < 2) {
    errors.lastName = 'Last name must have at least 2 characters.';
  }

  if (values?.email.length === 0) {
    errors.email = 'Email must be provided.';
  }

  if (values?.email && !verifyEmailAddress(values?.email)) {
    errors.email = 'Email must be valid.';
  }

  if (values?.password.length < 8) {
    errors.password = 'Password must be at least 8 characters long.';
  }

  if (values?.confirmPassword !== values?.password) {
    errors.confirmPassword = 'Password confirmation must match.';
  }

  if (values?.password.length > 71) {
    errors.password = 'Password is too long, must be less than 72 characters.';
  }

  return errors;
}

export function RegisterSuccessful(): React.JSX.Element {
  return (
    <div className={styles.registerPageRoot}>
      <MLogo className={layoutVariants({ size: 'logo' })} />
      <Typography align='center' className={styles.message} size='xl' weight='medium'>
        A verification message has been sent to your email address, please verify your email.
      </Typography>
      <div className={styles.footerRow}>
        <Typography color='subtle' size='sm'>
          Return to
        </Typography>
        <TextLink size='sm' to='/login'>
          Sign in
        </TextLink>
      </div>
    </div>
  );
}

export default function Register(): React.JSX.Element {
  const { enqueueSnackbar } = useSnackbar();
  const { data: config } = useAppConfiguration();
  const signUp = useSignUp();
  const [, navigate] = useLocation();
  const queryClient = useQueryClient();
  const [successful, setSuccessful] = useState(false);
  const pow = useProofOfWork('register', Boolean(config?.proofOfWorkEnabled));

  async function submit(values: RegisterValues, helpers: FormikHelpers<RegisterValues>): Promise<void> {
    helpers.setSubmitting(true);

    // null when disabled (challenge/nonce drop off). Kept in the chain so a
    // fetch/solve failure still hits the catch and finally below.
    return pow
      .getSolution()
      .then(solution =>
        signUp({
          betaCode: values.betaCode ?? null,
          captcha: values.captcha ?? null,
          email: values.email,
          firstName: values.firstName,
          lastName: values.lastName,
          password: values.password,
          timezone: getTimezone(),
          locale: getLocale(),
          challenge: solution?.challenge,
          nonce: solution?.nonce,
        }),
      )
      .then((result: SignUpResponse) => {
        // After sending the sign up request, if the user needs to verify their email then the requires verification
        // field will be true. We can stop here and just show the user a successful screen.
        if (result.requireVerification) {
          return setSuccessful(true);
        }

        return queryClient.invalidateQueries({ queryKey: ['/api/users/me'] }).then(() => {
          // If the register endpoint has told us to navigate to a specific url afterwards, then do that now.
          if (result.nextUrl) {
            return navigate(result.nextUrl);
          }

          // Otherwise just go to the index-ish route for the authenticated app.
          return navigate('/');
        });
      })
      .catch((error: ApiError<APIError>) => {
        // Single use, so line up a fresh one for a retry.
        pow.reset();
        const message =
          error?.response?.status === 429
            ? 'Too many requests, please try again in a few minutes'
            : error?.response?.data?.error || 'Failed to sign up.';
        enqueueSnackbar(message, {
          variant: 'error',
          disableWindowBlurListener: true,
        });
      })
      .finally(() => helpers.setSubmitting(false));
  }

  if (successful) {
    return <RegisterSuccessful />;
  }

  return (
    <div className={styles.registerPageRoot}>
      <MForm className={styles.form} initialValues={initialValues} onSubmit={submit} validate={validator}>
        <div className={styles.logo}>
          <MLogo />
        </div>
        <Flex align='center' orientation='column'>
          <Typography align='center' size='5xl'>
            Get Started
          </Typography>
          <Typography align='center' color='subtle' size='lg'>
            Create your monetr account now
          </Typography>
        </Flex>
        <Flex gap='sm' orientation='stackSmall'>
          <FormTextField
            autoFocus
            className={styles.input}
            data-testid='register-first-name'
            label='First Name'
            name='firstName'
            required
            type='text'
          />
          <FormTextField
            className={styles.input}
            data-testid='register-last-name'
            label='Last Name'
            name='lastName'
            required
            type='text'
          />
        </Flex>
        <FormTextField
          autoComplete='email'
          className={styles.input}
          data-testid='register-email'
          label='Email Address'
          name='email'
          required
          type='email'
        />
        <FormTextField
          autoComplete='new-password'
          className={styles.input}
          data-testid='register-password'
          label='Password'
          name='password'
          required
          type='password'
        />
        <FormTextField
          autoComplete='new-password'
          className={styles.input}
          data-testid='register-confirm-password'
          label='Confirm Password'
          name='confirmPassword'
          required
          type='password'
        />
        <BetaCodeInput />
        <MCaptcha className={styles.captcha} name='captcha' show={Boolean(config?.verifyRegister)} />
        <FormButton className={styles.submit} data-testid='register-submit' role='form' type='submit' variant='primary'>
          Sign Up
        </FormButton>
        <div className={styles.footerRow}>
          <Typography className={styles.termsBlurb} color='subtle' component='p' size='sm'>
            By signing up you agree to monetr's&nbsp;
            <a
              className={styles.inlineLink}
              href='https://monetr.app/policy/terms'
              rel='noopener noreferrer'
              target='_blank'
            >
              Terms & Conditions
            </a>{' '}
            and&nbsp;
            <a
              className={styles.inlineLink}
              href='https://monetr.app/policy/privacy'
              rel='noopener noreferrer'
              target='_blank'
            >
              Privacy Policy
            </a>
          </Typography>
        </div>
        <div className={styles.footerRow}>
          <Typography color='subtle' size='sm'>
            Already have an account?
          </Typography>
          <TextLink size='sm' to='/login'>
            Sign in instead
          </TextLink>
        </div>
      </MForm>
    </div>
  );
}
