import React from 'react';
import { useSelector } from 'react-redux';
import { Link as RouterLink } from 'react-router-dom';

import { getAllowForgotPassword } from 'shared/bootstrap/selectors';

export default function ForgotPasswordMaybe(): JSX.Element {
  const allowForgotPassword = useSelector(getAllowForgotPassword);

  if (!allowForgotPassword) {
    return null;
  }

  return (
    <div className="w-full flex justify-end mt-2.5 text-sm">
      <RouterLink className="opacity-50 hover:underline" to="/password/forgot">Forgot Password?</RouterLink>
    </div>
  );
}
