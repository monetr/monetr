import React from 'react';
import { Link as RouterLink } from 'react-router-dom';

import { useAppConfiguration } from 'hooks/useAppConfiguration';

export default function ForgotPasswordMaybe(): JSX.Element {
  const {
    allowForgotPassword,
  } = useAppConfiguration();

  if (!allowForgotPassword) {
    return null;
  }

  return (
    <div className="w-full flex justify-end mt-2.5 text-sm">
      <RouterLink className="opacity-50 hover:underline" to="/password/forgot">Forgot Password?</RouterLink>
    </div>
  );
}
