import LoginPage from 'pages/login';
import ForgotPasswordPage from 'pages/password/forgot';
import ResetPasswordPage from 'pages/password/reset';
import RegisterPage from 'pages/register';
import VerifyEmailPage from 'pages/verify/email';
import ResendVerificationPage from 'pages/verify/email/resend';
import React from 'react';
import { useSelector } from 'react-redux';
import { Navigate, Route, Routes } from 'react-router-dom';
import { getAllowForgotPassword, getSignUpAllowed } from 'shared/bootstrap/selectors';
import TOTPView from 'views/Authentication/TOTPView';

const UnauthenticatedApplication = (): JSX.Element => {
  const allowSignUp = useSelector(getSignUpAllowed);
  const allowForgotPassword = useSelector(getAllowForgotPassword);

  return (
    <Routes>
      <Route path="/login" element={ <LoginPage/> }/>
      <Route path="/login/mfa" element={ <TOTPView/> }/>
      { allowSignUp && <Route path="/register" element={ <RegisterPage/> }/> }
      { allowForgotPassword && <Route path="/password/forgot" element={ <ForgotPasswordPage/> }/> }
      { allowForgotPassword && <Route path="/password/reset" element={ <ResetPasswordPage/> }/> }
      <Route path="/verify/email" element={ <VerifyEmailPage/> }/>
      <Route path="/verify/email/resend" element={ <ResendVerificationPage/> }/>
      <Route path="*" element={ <Navigate replace to="/login"/> }/>
    </Routes>
  );
};

export default UnauthenticatedApplication;
