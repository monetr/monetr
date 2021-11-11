import React from 'react';
import { useSelector } from 'react-redux';
import { Redirect, Route, Switch } from 'react-router-dom';
import { getSignUpAllowed } from 'shared/bootstrap/selectors';
import LoginView from 'views/Authentication/LoginView';
import ResendVerification from 'views/Authentication/ResendVerification';
import SignUpView from 'views/Authentication/SignUpView';
import VerifyEmail from 'views/Authentication/VerifyEmail';

const UnauthenticatedApplication = (): JSX.Element => {
  const allowSignUp = useSelector(getSignUpAllowed);

  return (
    <Switch>
      <Route path="/login">
        <LoginView/>
      </Route>
      { allowSignUp &&
      <Route path="/register">
        <SignUpView/>
      </Route>
      }
      <Route path="/verify/email" exact component={ VerifyEmail }/>
      <Route path="/verify/email/resend" exact component={ ResendVerification }/>
      <Route>
        <Redirect to={ { pathname: '/login' } }/>
      </Route>
    </Switch>
  );
};

export default UnauthenticatedApplication;