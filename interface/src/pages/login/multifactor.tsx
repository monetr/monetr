import React from 'react';
import { FormikHelpers } from 'formik';

import { InputOTP, InputOTPGroup, InputOTPSeparator, InputOTPSlot } from '@monetr/interface/components/InputOTP';
import MFormButton from '@monetr/interface/components/MButton';
import MForm from '@monetr/interface/components/MForm';
import MLogo from '@monetr/interface/components/MLogo';
import MSpan from '@monetr/interface/components/MSpan';
import request from '@monetr/interface/util/request';

interface MultifactorValues {
  totp: string;
}

const initialValues: MultifactorValues = {
  totp: '',
};

export default function MultifactorAuthenticationPage(): JSX.Element {

  async function submit(values: MultifactorValues, helpers: FormikHelpers<MultifactorValues>) {
    helpers.setSubmitting(true);

    return request().post('/authentication/multifactor', {
      totp: values.totp,
    })
      .then(result => {
        console.log(result);
      })
      .catch(error => {
        console.log(error);
      });
  }

  return (
    <MForm
      initialValues={ initialValues }
      onSubmit={ submit }
      className='w-full h-full flex pt-10 md:pt-0 md:pb-10 md:justify-center items-center flex-col gap-1 px-5'
    >
      <div className='max-w-[128px] w-full'>
        <MLogo />
      </div>
      <MSpan>Please provide your one time password from your authenticator app</MSpan>
      <InputOTP name='totp' maxLength={ 6 }>
        <InputOTPGroup>
          <InputOTPSlot index={ 0 } />
          <InputOTPSlot index={ 1 } />
          <InputOTPSlot index={ 2 } />
        </InputOTPGroup>
        <InputOTPSeparator />
        <InputOTPGroup>
          <InputOTPSlot index={ 3 } />
          <InputOTPSlot index={ 4 } />
          <InputOTPSlot index={ 5 } />
        </InputOTPGroup>
      </InputOTP>
      <div className='w-full xl:w-1/5 lg:w-1/4 md:w-1/3 sm:w-1/2 mt-1'>
        <MFormButton
          data-testid='login-submit'
          color='primary'
          variant='solid'
          role='form'
          type='submit'
          className='w-full'
          tabIndex={ 3 }
        >
          Continue
        </MFormButton>
      </div>
    </MForm>
  );
}
