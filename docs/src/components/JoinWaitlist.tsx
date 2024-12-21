/* eslint-disable max-len */

import { useCallback, useState } from 'react';
import { Field, Form, Formik, FormikHelpers } from 'formik';

import { z } from 'zod';
import { toFormikValidationSchema } from 'zod-formik-adapter';

const waitlistForm = z.object({
  email: z.string().email(),
});

type WaitlistForm = z.infer<typeof waitlistForm>;

export default function JoinWaitlist(): JSX.Element {
  const [submitted, setSubmitted] = useState(false);
  const onSubmit = useCallback((
    values: WaitlistForm,
    helpers: FormikHelpers<WaitlistForm>,
  ) => {
    helpers.setSubmitting(true);
    const data = new FormData();
    data.append('email', values.email);
    data.append('l', 'abd7e09e-42d5-40c6-9a84-daaf77dc4d5f');
    fetch('https://newsletter.monetr.app/subscription/form', {
      body: data,
      method: 'POST',
    })
      .then(() => setSubmitted(true))
      .finally(() => helpers.setSubmitting(false));
  }, [setSubmitted]);

  if (submitted) {
    return (
      <div className='w-full flex flex-col items-center'>
        <h2 className='sm:text-lg font-medium'>
          Thank you! You'll be notified when monetr is available!
        </h2>
      </div>
    );
  }

  return (
    <div className='w-full flex flex-col items-center'>
      <Formik<WaitlistForm>
        onSubmit={ onSubmit }
        initialValues={ { email: '' } }
        validationSchema={ toFormikValidationSchema(waitlistForm) }
      >
        <Form className='max-w-xl w-full space-x-2 flex'>
          <label htmlFor='email-address' className='sr-only'>
            Email address
          </label>
          <Field
            id='email-address'
            name='email'
            type='email'
            autoComplete='email'
            required
            className='min-w-0 flex-auto rounded-md bg-white/5 px-3.5 py-2 text-base text-white outline outline-1 -outline-offset-1 outline-white/10 placeholder:text-gray-500 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-purple-500 sm:text-sm/6'
            placeholder='Enter your email, we will let you know when we launch!'
          />
          <button
            type='submit'
            className='flex-none rounded-md bg-purple-500 px-3.5 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-purple-400 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-purple-500'
          >
            Get Notified
          </button>
        </Form>
      </Formik>
    </div>
  );
}
