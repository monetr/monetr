import PropTypes from "prop-types";
import React, { PureComponent } from 'react';
import { connect } from "react-redux";
import { Redirect, Route, Switch, withRouter, } from "react-router-dom";
import { bindActionCreators } from "redux";
import bootstrapLogin from "shared/authentication/actions/bootstrapLogin";
import { getIsAuthenticated, getSubscriptionIsActive } from "shared/authentication/selectors";
import bootstrapApplication from "shared/bootstrap";
import { getIsBootstrapped, getSignUpAllowed } from "shared/bootstrap/selectors";
import LoginView from "views/Login";
import SignUpView from "views/SignUp";
import { Backdrop, CircularProgress } from "@material-ui/core";
import { BillingRequired } from "views/Subscriptions/BillingRequired";
import AuthenticatedApp from "AuthenticatedApp";

export class Root extends PureComponent {
  state = {
    loading: true,
  };

  static propTypes = {
    history: PropTypes.object.isRequired,
    isReady: PropTypes.bool.isRequired,
    isAuthenticated: PropTypes.bool.isRequired,
    isSubscriptionActive: PropTypes.bool.isRequired,
    allowSignUp: PropTypes.bool.isRequired,
    bootstrapApplication: PropTypes.func.isRequired,
    bootstrapLogin: PropTypes.func.isRequired,
  };

  componentDidMount() {
    this.attemptBootstrap();
  }

  attemptBootstrap = () => {
    this.props.bootstrapApplication()
      .then(() => {
        return this.props.bootstrapLogin()
          .then(result => {
            if (result && result.data.nextUrl) {
              this.props.history.push(result.data.nextUrl);
            }
          });
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
      return <BillingRequired />;
    }

    return (
      <AuthenticatedApp />
    )
  }
}

export default connect(
  state => ({
    isAuthenticated: getIsAuthenticated(state),
    isReady: getIsBootstrapped(state),
    isSubscriptionActive: getSubscriptionIsActive(state),
    allowSignUp: getSignUpAllowed(state),
  }),
  dispatch => bindActionCreators({
    bootstrapApplication,
    bootstrapLogin,
  }, dispatch),
)(withRouter(Root));
