import { useState } from 'react';
import type { FormikErrors, FormikHelpers } from 'formik';

import FormButton from '@monetr/interface/components/FormButton';
import FormTextField from '@monetr/interface/components/FormTextField';
import MForm from '@monetr/interface/components/MForm';
import MLink from '@monetr/interface/components/MLink';
import MLogo from '@monetr/interface/components/MLogo';
import Typography from '@monetr/interface/components/Typography';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import { useProofOfWork } from '@monetr/interface/hooks/useProofOfWork';
import useSendForgotPassword from '@monetr/interface/hooks/useSendForgotPassword';
import verifyEmailAddress from '@monetr/interface/util/verifyEmailAddress';

import styles from './forgot.module.scss';

interface Values {
  email: string;
}

const initialValues: Values = {
  email: '',
};

export function ForgotPasswordComplete(): React.JSX.Element {
  return (
    <div className={styles.root}>
      <div className={styles.logo}>
        <MLogo />
      </div>
      <div className={styles.header}>
        <Typography size='inherit'>Check your email</Typography>
        <Typography className={styles.completeMessage} color='subtle' size='sm'>
          If a user was found with the email provided, then you should receive an email with instructions on how to
          reset your password.
        </Typography>
      </div>
      <div className={styles.signInRow}>
        <Typography color='subtle' size='sm'>
          Return to
        </Typography>
        <MLink size='sm' to='/login'>
          Sign in
        </MLink>
      </div>
    </div>
  );
}

export default function ForgotPasswordNew(): React.JSX.Element {
  const { data: config } = useAppConfiguration();
  const sendForgotPassword = useSendForgotPassword();
  const [isComplete, setIsComplete] = useState<boolean>(false);
  const pow = useProofOfWork('forgot', Boolean(config?.proofOfWorkEnabled));

  function validate(values: Values): FormikErrors<Values> {
    const errors: FormikErrors<Values> = {};

    if (values.email && !verifyEmailAddress(values.email)) {
      errors.email = 'Please provide a valid email address.';
    }

    return errors;
  }

  async function submit(values: Values, helpers: FormikHelpers<Values>): Promise<void> {
    helpers.setSubmitting(true);

    // sendForgotPassword does the work and shows its own errors; we just clear
    // submitting. getSolution is null when disabled.
    return await pow
      .getSolution()
      .then(solution =>
        sendForgotPassword({
          email: values.email,
          challenge: solution?.challenge,
          nonce: solution?.nonce,
        }),
      )
      .then(() => setIsComplete(true))
      .catch(() => {
        // A rejection here is the proof of work (the hook swallows its own
        // errors), so line up a fresh challenge.
        pow.reset();
      })
      .finally(() => helpers.setSubmitting(false));
  }

  if (isComplete) {
    return <ForgotPasswordComplete />;
  }

  return (
    <MForm
      className={styles.root}
      initialValues={initialValues}
      onInput={pow.warmup}
      onSubmit={submit}
      validate={validate}
    >
      <div className={styles.logo}>
        <MLogo />
      </div>
      <div className={styles.header}>
        <Typography size='inherit'>Forgot your password?</Typography>
        <Typography color='subtle' size='sm'>
          We can email you a link to reset it.
        </Typography>
      </div>
      <FormTextField
        autoComplete='username'
        autoFocus
        className={styles.input}
        data-testid='forgot-email'
        label='Email Address'
        name='email'
        required
        type='email'
      />
      <div className={styles.submitWrapper}>
        <FormButton className={styles.button} data-testid='forgot-submit' role='form' type='submit' variant='primary'>
          Reset Password
        </FormButton>
      </div>
      <div className={styles.signInRow}>
        <Typography color='subtle' size='sm'>
          Remembered your password?
        </Typography>
        <MLink size='sm' to='/login'>
          Sign in
        </MLink>
      </div>
    </MForm>
  );
}
