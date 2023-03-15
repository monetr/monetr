import React, { lazy } from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';

import { useAppConfiguration } from 'hooks/useAppConfiguration';
import LoginPage from 'pages/login';

const ForgotPasswordPage = lazy(() => import('pages/password/forgot'));
const ResetPasswordPage = lazy(() => import('pages/password/reset'));
const RegisterPage = lazy(() => import('pages/register'));
const VerifyEmailPage = lazy(() => import('pages/verify/email'));
const ResendVerificationPage = lazy(() => import('pages/verify/email/resend'));
const TOTPView = lazy(() => import('views/Authentication/TOTPView'));

export default function UnauthenticatedApplication(): JSX.Element {
  const {
    allowSignUp,
    allowForgotPassword,
  } = useAppConfiguration();

  return (
    <Routes>
      <Route path="/login" element={ <LoginPage /> } />
      <Route path="/login/mfa" element={ <TOTPView /> } />
      { allowSignUp && <Route path="/register" element={ <RegisterPage /> } /> }
      { allowForgotPassword && <Route path="/password/forgot" element={ <ForgotPasswordPage /> } /> }
      <Route path="/password/reset" element={ <ResetPasswordPage /> } />
      <Route path="/verify/email" element={ <VerifyEmailPage /> } />
      <Route path="/verify/email/resend" element={ <ResendVerificationPage /> } />
      <Route path="*" element={ <Navigate replace to="/login" /> } />
    </Routes>
  );
}
