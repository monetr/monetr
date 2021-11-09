import React, { Component } from 'react';
import { connect } from 'react-redux';
import { Redirect, Route, RouteComponentProps, Switch, withRouter, } from 'react-router-dom';
import bootstrapLogin from 'shared/authentication/actions/bootstrapLogin';
import { getIsAuthenticated, getSubscriptionIsActive } from 'shared/authentication/selectors';
import bootstrapApplication from 'shared/bootstrap/actions/bootstrapApplication';
import { getIsBootstrapped, getSignUpAllowed } from 'shared/bootstrap/selectors';
import { AppState } from 'store';
import LoginView from 'views/Authentication/LoginView';
import SignUpView from 'views/Authentication/SignUpView';
import { Backdrop, CircularProgress } from '@mui/material';
import BillingRequiredRouter from 'views/Subscriptions/BillingRequiredRouter';
import AuthenticatedApp from 'AuthenticatedApp';
import VerifyEmail from 'views/Authentication/VerifyEmail';
import ResendVerification from 'views/Authentication/ResendVerification';

interface WithConnectionPropTypes {
  isReady: boolean;
  isAuthenticated: boolean;
  isSubscriptionActive: boolean;
  allowSignUp: boolean;
  bootstrapApplication: () => Promise<void>;
  bootstrapLogin: () => Promise<void>;
}

interface State {
  loading: boolean;
}

export class Root extends Component<RouteComponentProps & WithConnectionPropTypes, State> {
  state = {
    loading: true,
  };

  componentDidMount() {
    this.attemptBootstrap();
  }

  attemptBootstrap = () => {
    const { bootstrapApplication, bootstrapLogin } = this.props;

    bootstrapApplication()
      .then(() => {
        return bootstrapLogin()
      })
      .catch(error => {
        throw error;
      })
      .finally(() => {
        this.setState({
          loading: false
        });
      });
  };

  renderUnauthenticated = () => {
    return (
      <Switch>
        <Route path="/login">
          <LoginView/>
        </Route>
        { this.props.allowSignUp &&
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
    )
  };

  render() {
    const { isReady, isAuthenticated, isSubscriptionActive } = this.props;
    if (!isReady || this.state.loading) {
      return (
        <Backdrop open={ true }>
          <CircularProgress color="inherit"/>
        </Backdrop>
      );
    }

    if (!isAuthenticated) {
      return this.renderUnauthenticated();
    }

    if (!isSubscriptionActive) {
      return <BillingRequiredRouter/>;
    }

    return (
      <AuthenticatedApp/>
    )
  }
}

export default connect(
  (state: AppState) => ({
    isAuthenticated: getIsAuthenticated(state),
    isReady: getIsBootstrapped(state),
    isSubscriptionActive: getSubscriptionIsActive(state),
    allowSignUp: getSignUpAllowed(state),
  }),
  {
    bootstrapApplication,
    bootstrapLogin,
  },
)(withRouter(Root));
