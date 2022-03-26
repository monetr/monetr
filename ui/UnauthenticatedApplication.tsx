import React from 'react';
import { useSelector } from 'react-redux';
import { Navigate, Route, Routes } from 'react-router-dom';
import { getAllowForgotPassword, getSignUpAllowed } from 'shared/bootstrap/selectors';
import ForgotPasswordView from 'views/Authentication/ForgotPasswordView';
import LoginView from 'views/Authentication/LoginView';
import ResendVerification from 'views/Authentication/ResendVerification';
import ResetPasswordView from 'views/Authentication/ResetPasswordView';
import SignUpView from 'views/Authentication/SignUpView';
import TOTPView from 'views/Authentication/TOTPView';
import VerifyEmail from 'views/Authentication/VerifyEmail';

const UnauthenticatedApplication = (): JSX.Element => {
  const allowSignUp = useSelector(getSignUpAllowed);
  const allowForgotPassword = useSelector(getAllowForgotPassword);

  return (
    <Routes>
      <Route path="/login" element={ <LoginView/> }/>
      <Route path="/login/mfa" element={ <TOTPView/> }/>
      { allowSignUp && <Route path="/register" element={ <SignUpView/> }/> }
      { allowForgotPassword && <Route path="/password/forgot" element={ <ForgotPasswordView/> }/> }
      { allowForgotPassword && <Route path="/password/reset" element={ <ResetPasswordView/> }/> }
      <Route path="/verify/email" element={ <VerifyEmail/> }/>
      <Route path="/verify/email/resend" element={ <ResendVerification/> }/>
      <Route path="*" element={ <Navigate replace to="/login"/> }/>
    </Routes>
  );
};

export default UnauthenticatedApplication;
