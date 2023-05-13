import React from 'react';

import MButton from 'components/MButton';
import MLink from 'components/MLink';
import MLogo from 'components/MLogo';
import MSpan from 'components/MSpan';
import MTextField from 'components/MTextField';
import { useAppConfiguration } from 'hooks/useAppConfiguration';

export default function LoginNew(): JSX.Element {
  const config = useAppConfiguration();

  function ForgotPasswordButton(): JSX.Element {
    // If the application is not configured to allow forgot password then don't show the button.
    if (!config.allowForgotPassword) {
      return null;
    }

    return (
      <div className="text-sm">
        <MLink to="/forgot">
          Forgot password?
        </MLink>
      </div>
    );
  }

  function SignUpButton(): JSX.Element {
    if (!config.allowSignUp) return null;

    return (
      <div className="w-full lg:w-1/4 sm:w-1/3 mt-1 flex justify-center gap-1 text-sm">
        <MSpan variant="light">Not a user?</MSpan>
        <MLink to="/register">Sign up now</MLink>
      </div>
    );
  }

  return (
    <form className="w-full h-full flex pt-10 md:pt-0 md:pb-10 md:justify-center items-center flex-col gap-5">
      <div className="max-w-[128px] w-full">
        <MLogo />
      </div>
      <MSpan>Sign into your monetr account</MSpan>
      <MTextField
        label="Email Address"
        name='email'
        type='email'
        required
        className="w-full lg:w-1/4 sm:w-1/3"
      />
      <MTextField
        label="Password"
        name='password'
        type='password'
        required
        labelDecorator={ ForgotPasswordButton }
        className="w-full lg:w-1/4 sm:w-1/3"
      />
      <div className="w-full lg:w-1/4 sm:w-1/3 mt-1">
        <MButton color="primary" variant="solid">
          Sign In
        </MButton>
      </div>
      <SignUpButton />
    </form>
  );
}

