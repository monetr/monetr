import React, { Fragment } from 'react';

import MButton from 'components/MButton';
import MCheckbox from 'components/MCheckbox';
import MForm from 'components/MForm';
import MLink from 'components/MLink';
import MLogo from 'components/MLogo';
import MSpan from 'components/MSpan';
import MTextField from 'components/MTextField';
import { useAppConfiguration } from 'hooks/useAppConfiguration';

export default function RegisterNew(): JSX.Element {
  const config = useAppConfiguration();

  function BetaCodeInput(): JSX.Element {
    if (!config.requireBetaCode) return null;

    return (
      <MTextField
        label="Beta Code"
        name="betaCode"
        type="text"
        required
        uppercase
        className="xl:w-2/5 lg:w-1/3 md:w-1/2 w-full"
      />
    );
  }

  return (
    <MForm className="w-full h-full flex pt-10 md:pt-0 md:pb-10 md:justify-center items-center flex-col gap-5 px-5">
      <div className="max-w-[96px] w-full">
        <MLogo />
      </div>
      <div className="flex flex-col items-center">
        <MSpan size="5xl">
          Get Started
        </MSpan>
        <MSpan size="lg" variant="light">
          Create your monetr account now
        </MSpan>
      </div>
      <div className="grid grid-flow-row sm:grid-flow-col gap-5 xl:w-2/5 lg:w-1/3 md:w-1/2 w-full">
        <MTextField
          autoFocus
          label="First Name"
          name="firstName"
          type="text"
          required
          className="w-full"
        />
        <MTextField
          label="Last Name"
          name="lastName"
          type="text"
          required
          className="w-full"
        />
      </div>
      <MTextField
        label="Email Address"
        name='email'
        type='email'
        required
        className="xl:w-2/5 lg:w-1/3 md:w-1/2 w-full"
      />
      <MTextField
        label="Password"
        name='password'
        type='password'
        required
        className="xl:w-2/5 lg:w-1/3 md:w-1/2 w-full"
      />
      <MTextField
        label="Confirm Password"
        name='confirmPassword'
        type='password'
        required
        className="xl:w-2/5 lg:w-1/3 md:w-1/2 w-full"
      />
      <BetaCodeInput />
      <MCheckbox
        id="terms"
        name="agree"
        label={
          <Fragment>
            I agree to monetr's&nbsp;
            <a
              target="_blank"
              className="text-blue-500 hover:underline focus:ring-2 focus:ring-blue-500 focus:underline"
              href='https://github.com/monetr/legal/blob/main/TERMS_OF_USE.md'>
              Terms of Use
            </a> and&nbsp;
            <a
              target="_blank"
              className="text-blue-500 hover:underline focus:ring-2 focus:ring-blue-500 focus:underline"
              href='https://github.com/monetr/legal/blob/main/PRIVACY.md'
            >
              Privacy Policy
            </a>
          </Fragment>
        }
      />
      <div className="xl:w-2/5 lg:w-1/3 md:w-1/2 w-full mt-1">
        <MButton color="primary" variant="solid" role="form" type="submit">
          Sign Up
        </MButton>
      </div>
      <div className="xl:w-2/5 lg:w-1/3 md:w-1/2 w-full mt-1 flex justify-center gap-1">
        <MSpan variant="light" size="sm">Already have an account?</MSpan>
        <MLink to="/login" size="sm">Sign in instead</MLink>
      </div>
    </MForm>
  );
}
