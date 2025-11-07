import type { FormikErrors, FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import FormButton from '@monetr/interface/components/FormButton';
import FormTextField from '@monetr/interface/components/FormTextField';
import MCaptcha from '@monetr/interface/components/MCaptcha';
import MForm from '@monetr/interface/components/MForm';
import MLogo from '@monetr/interface/components/MLogo';
import TextLink from '@monetr/interface/components/TextLink';
import Typography from '@monetr/interface/components/Typography';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import useLogin from '@monetr/interface/hooks/useLogin';
import verifyEmailAddress from '@monetr/interface/util/verifyEmailAddress';

import styles from './login.module.scss';

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
  const errors: FormikErrors<LoginValues> = {};

  if (values?.email.length === 0) {
    errors.email = 'Email must be provided.';
  }

  if (values?.email && !verifyEmailAddress(values?.email)) {
    errors.email = 'Email must be valid.';
  }

  if (values?.password.length < 8) {
    errors.password = 'Password must be at least 8 characters long.';
  }

  return errors;
}

export default function Login(): JSX.Element {
  const { enqueueSnackbar } = useSnackbar();
  const { data: config } = useAppConfiguration();
  const login = useLogin();

  async function submit(values: LoginValues, helpers: FormikHelpers<LoginValues>) {
    helpers.setSubmitting(true);

    return login({
      captcha: values.captcha,
      email: values.email,
      password: values.password,
    })
      .catch(error =>
        enqueueSnackbar(error?.response?.data?.error || 'Failed to authenticate.', {
          variant: 'error',
          disableWindowBlurListener: true,
        }),
      )
      .finally(() => helpers.setSubmitting(false));
  }

  return (
    <MForm initialValues={initialValues} validate={validator} onSubmit={submit} className={styles.root}>
      <div className={styles.logo}>
        <MLogo />
      </div>
      <Typography>Sign into your monetr account</Typography>
      <FormTextField
        data-testid='login-email'
        autoFocus
        label='Email Address'
        name='email'
        type='email'
        required
        className={styles.input}
      />
      <FormTextField
        autoComplete='current-password'
        className={styles.input}
        data-testid='login-password'
        label='Password'
        labelDecorator={ForgotPasswordButton}
        name='password'
        required
        type='password'
      />
      <MCaptcha name='captcha' show={Boolean(config?.verifyLogin)} />
      <FormButton data-testid='login-submit' variant='primary' role='form' type='submit' className={styles.input}>
        Sign In
      </FormButton>
      {Boolean(config?.allowSignUp) && (
        <div className={styles.signUpWrapper}>
          <Typography size='sm' color='subtle'>
            Not a user?
          </Typography>
          <TextLink to='/register' size='sm' data-testid='login-signup'>
            Sign up now
          </TextLink>
        </div>
      )}
    </MForm>
  );
}

function ForgotPasswordButton(): JSX.Element {
  const { data: config } = useAppConfiguration();
  // If the application is not configured to allow forgot password then don't show the button.
  if (!config?.allowForgotPassword) {
    return null;
  }

  return (
    <TextLink to='/password/forgot' size='sm' data-testid='login-forgot' tabIndex={-1}>
      Forgot password?
    </TextLink>
  );
}
