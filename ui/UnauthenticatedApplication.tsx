import React from 'react';
import { useSelector } from 'react-redux';
import { Navigate, Route, Routes } from 'react-router-dom';
import { getSignUpAllowed } from 'shared/bootstrap/selectors';
import LoginView from 'views/Authentication/LoginView';
import ResendVerification from 'views/Authentication/ResendVerification';
import SignUpView from 'views/Authentication/SignUpView';
import VerifyEmail from 'views/Authentication/VerifyEmail';

const UnauthenticatedApplication = (): JSX.Element => {
  const allowSignUp = useSelector(getSignUpAllowed);

  return (
    <Routes>
      <Route path="/login" element={ <LoginView/> }/>
      { allowSignUp && <Route path="/register" element={ <SignUpView/> }/> }
      <Route path="/verify/email" element={ <VerifyEmail/> }/>
      <Route path="/verify/email/resend" element={ <ResendVerification/> }/>
      <Route path="*" element={ <Navigate replace to="/login"/> }/>
    </Routes>
  );
};

export default UnauthenticatedApplication;