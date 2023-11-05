import React, { useState } from 'react';
import { FormikErrors, FormikHelpers } from 'formik';

import MFormButton from '@monetr/interface/components/MButton';
import MCaptcha from '@monetr/interface/components/MCaptcha';
import MForm from '@monetr/interface/components/MForm';
import MLink from '@monetr/interface/components/MLink';
import MLogo from '@monetr/interface/components/MLogo';
import MSpan from '@monetr/interface/components/MSpan';
import MTextField from '@monetr/interface/components/MTextField';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import useSendForgotPassword from '@monetr/interface/hooks/useSendForgotPassword';
import verifyEmailAddress from '@monetr/interface/util/verifyEmailAddress';

interface Values {
  email: string;
  captcha: string | null;
}

const initialValues: Values = {
  email: '',
  captcha: null,
};

export function ForgotPasswordComplete(): JSX.Element {
  return (
    <div className="w-full h-full flex pt-10 md:pt-0 md:pb-10 md:justify-center items-center flex-col gap-1 px-5">
      <div className="max-w-[128px] w-full">
        <MLogo />
      </div>
      <div className="flex flex-col items-center">
        <MSpan>
          Check your email
        </MSpan>
        <MSpan color="subtle" className="max-w-[248px] text-center text-sm">
          If a user was found with the email provided, then you should receive an email with instructions on how to
          reset your password.
        </MSpan>
      </div>
      <div className="w-full lg:w-1/4 sm:w-1/3 mt-1 flex justify-center gap-1">
        <MSpan color="subtle" className='text-sm'>Return to</MSpan>
        <MLink to="/login" size="sm">Sign in</MLink>
      </div>
    </div>
  );
}

export default function ForgotPasswordNew(): JSX.Element {
  const config = useAppConfiguration();
  const sendForgotPassword = useSendForgotPassword();
  const [isComplete, setIsComplete] = useState<boolean>(false);

  function validate(values: Values): FormikErrors<Values> {
    const errors: FormikErrors<Values> = {};

    if (values.email && !verifyEmailAddress(values.email)) {
      errors['email'] = 'Please provide a valid email address.';
    }

    return errors;
  }

  async function submit(values: Values, helpers: FormikHelpers<Values>): Promise<void> {
    helpers.setSubmitting(true);

    // sendForgotPassword pretty much does all the work, the only thing we need to do is make sure that once we are done
    // we set submitting back to false.
    return sendForgotPassword(values.email, values.captcha)
      .then(() => setIsComplete(true))
      .finally(() => helpers.setSubmitting(false));
  }

  if (isComplete) {
    return (
      <ForgotPasswordComplete />
    );
  }

  return (
    <MForm
      initialValues={ initialValues }
      validate={ validate }
      onSubmit={ submit }
      className="w-full h-full flex pt-10 md:pt-0 md:pb-10 md:justify-center items-center flex-col gap-1 px-5"
    >
      <div className="max-w-[128px] w-full">
        <MLogo />
      </div>
      <div className="flex flex-col items-center">
        <MSpan>
          Forgot your password?
        </MSpan>
        <MSpan color="subtle" className='text-sm'>
          We can email you a link to reset it.
        </MSpan>
      </div>
      <MTextField
        autoFocus
        autoComplete='username'
        label="Email Address"
        name='email'
        type='email'
        required
        className="w-full xl:w-1/5 lg:w-1/4 md:w-1/3 sm:w-1/2"
      />
      <MCaptcha
        className='mb-1'
        name="captcha"
        show={ Boolean(config?.verifyForgotPassword) }
      />
      <div className="w-full xl:w-1/5 lg:w-1/4 md:w-1/3 sm:w-1/2 mt-1">
        <MFormButton
          color="primary"
          variant="solid"
          role="form"
          type="submit"
          className='w-full'
        >
          Reset Password
        </MFormButton>
      </div>
      <div className="w-full lg:w-1/4 sm:w-1/3 mt-1 flex justify-center gap-1">
        <MSpan color="subtle" className='text-sm'>Remembered your password?</MSpan>
        <MLink to="/login" size="sm">Sign in</MLink>
      </div>
    </MForm>
  );
}
