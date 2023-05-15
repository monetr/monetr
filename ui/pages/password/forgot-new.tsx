import React from 'react';
import { Formik, FormikErrors, FormikHelpers } from 'formik';

import MButton from 'components/MButton';
import MForm from 'components/MForm';
import MLink from 'components/MLink';
import MLogo from 'components/MLogo';
import MSpan from 'components/MSpan';
import MTextField from 'components/MTextField';
import useSendForgotPassword from 'hooks/useSendForgotPassword';
import verifyEmailAddress from 'util/verifyEmailAddress';

interface Values {
  email: string;
}

const initialValues: Values = {
  email: '',
};

export default function ForgotPasswordNew(): JSX.Element {
  const sendForgotPassword = useSendForgotPassword();

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
    // NOTE: The verification passed here is always null at the moment.
    return sendForgotPassword(values.email, null)
      .finally(() => helpers.setSubmitting(false));
  }

  return (
    <Formik initialValues={ initialValues } validate={ validate } onSubmit={ submit }>
      <MForm className="w-full h-full flex pt-10 md:pt-0 md:pb-10 md:justify-center items-center flex-col gap-5 px-5">
        <div className="max-w-[128px] w-full">
          <MLogo />
        </div>
        <div className="flex flex-col items-center">
          <MSpan>
            Forgot your password?
          </MSpan>
          <MSpan size="sm" variant="light">
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
        <div className="w-full xl:w-1/5 lg:w-1/4 md:w-1/3 sm:w-1/2 mt-1">
          <MButton
            color="primary"
            variant="solid"
            role="form"
            type="submit"
          >
            Reset Password
          </MButton>
        </div>
        <div className="w-full lg:w-1/4 sm:w-1/3 mt-1 flex justify-center gap-1">
          <MSpan variant="light" size="sm">Remembered your password?</MSpan>
          <MLink to="/login" size="sm">Sign in</MLink>
        </div>
      </MForm>
    </Formik>
  );
}
