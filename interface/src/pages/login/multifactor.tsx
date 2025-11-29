import { Fragment } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import type { FormikErrors, FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import FormButton from '@monetr/interface/components/FormButton';
import { InputOTP, InputOTPGroup, InputOTPSeparator, InputOTPSlot } from '@monetr/interface/components/InputOTP';
import UnauthenticatedLogo from '@monetr/interface/components/Layout/UnauthenticatedLogo';
import MForm from '@monetr/interface/components/MForm';
import LogoutFooter from '@monetr/interface/components/setup/LogoutFooter';
import Typography from '@monetr/interface/components/Typography';
import request from '@monetr/interface/util/request';

import styles from './multifactor.module.scss';

interface MultifactorValues {
  totp: string;
}

const initialValues: MultifactorValues = {
  totp: '',
};

export default function MultifactorAuthenticationPage(): JSX.Element {
  const { enqueueSnackbar } = useSnackbar();
  const queryClient = useQueryClient();

  async function submit(values: MultifactorValues, helpers: FormikHelpers<MultifactorValues>) {
    helpers.setSubmitting(true);

    return request()
      .post('/authentication/multifactor', {
        totp: values.totp,
      })
      .then(() => queryClient.invalidateQueries({ queryKey: ['/users/me'] }))
      .catch(error =>
        enqueueSnackbar(error?.response?.data?.error || 'Failed to validate TOTP code.', {
          variant: 'error',
          disableWindowBlurListener: true,
        }),
      );
  }

  function validate(values: MultifactorValues): FormikErrors<MultifactorValues> {
    const errors: FormikErrors<MultifactorValues> = {};
    if (values.totp.length !== 6) {
      errors.totp = 'TOTP code must be 6 digits';
    }

    return errors;
  }

  function formatContinueButton(values: MultifactorValues): string {
    const length = values.totp.length;
    if (length < 6) {
      return `${6 - length} digits left`;
    }

    return 'Continue';
  }

  return (
    <MForm
      className={styles.multifactorRoot}
      initialValues={initialValues}
      onSubmit={submit}
      validate={validate}
      validateOnMount
    >
      {formik => (
        <Fragment>
          <UnauthenticatedLogo />
          <Typography align='center' component='p'>
            Please provide the 6-digit code from your authenticator app
          </Typography>
          <InputOTP
            autoFocus
            disabled={formik.isSubmitting}
            maxLength={6}
            name='totp'
            onChange={value => formik.setFieldValue('totp', value)}
            required
          >
            <InputOTPGroup>
              <InputOTPSlot index={0} />
              <InputOTPSlot index={1} />
              <InputOTPSlot index={2} />
            </InputOTPGroup>
            <InputOTPSeparator />
            <InputOTPGroup>
              <InputOTPSlot index={3} />
              <InputOTPSlot index={4} />
              <InputOTPSlot index={5} />
            </InputOTPGroup>
          </InputOTP>
          <FormButton
            className={styles.input}
            data-testid='multifactor-submit'
            disabled={!formik.isValid}
            role='form'
            type='submit'
            variant='primary'
          >
            {formatContinueButton(formik.values)}
          </FormButton>
          <LogoutFooter />
        </Fragment>
      )}
    </MForm>
  );
}
