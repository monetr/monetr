import { Fragment } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import type { FormikErrors, FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import FormButton from '@monetr/interface/components/FormButton';
import { InputOTP, InputOTPGroup, InputOTPSeparator, InputOTPSlot } from '@monetr/interface/components/InputOTP';
import MForm from '@monetr/interface/components/MForm';
import MLogo from '@monetr/interface/components/MLogo';
import MSpan from '@monetr/interface/components/MSpan';
import LogoutFooter from '@monetr/interface/components/setup/LogoutFooter';
import request from '@monetr/interface/util/request';

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
      initialValues={initialValues}
      onSubmit={submit}
      validate={validate}
      validateOnMount
      className='w-full h-full flex pt-10 md:pt-0 md:pb-10 md:justify-center items-center flex-col gap-2 px-5'
    >
      {formik => (
        <Fragment>
          <div className='max-w-[128px] w-full'>
            <MLogo />
          </div>
          <MSpan className='w-full xl:w-1/5 lg:w-1/4 md:w-1/3 sm:w-1/2 mt-1 text-center'>
            Please provide the 6-digit code from your authenticator app
          </MSpan>
          <InputOTP
            name='totp'
            maxLength={6}
            required
            onChange={value => formik.setFieldValue('totp', value)}
            disabled={formik.isSubmitting}
            autoFocus
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
          <div className='w-full xl:w-1/5 lg:w-1/4 md:w-1/3 sm:w-1/2 mt-1'>
            <FormButton
              disabled={!formik.isValid}
              data-testid='multifactor-submit'
              variant='primary'
              role='form'
              type='submit'
              className='w-full'
              tabIndex='0'
            >
              {formatContinueButton(formik.values as MultifactorValues)}
            </FormButton>
          </div>
          <LogoutFooter />
        </Fragment>
      )}
    </MForm>
  );
}
