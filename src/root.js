import AuthenticatedApplication from "AuthenticatedApplication";
import PropTypes from "prop-types";
import React, { PureComponent } from 'react';
import { connect } from "react-redux";
import { BrowserRouter as Router, Redirect, Route, Switch, } from "react-router-dom";
import { bindActionCreators } from "redux";
import bootstrapLogin from "shared/authentication/actions/bootstrapLogin";
import { getIsAuthenticated } from "shared/authentication/selectors";
import bootstrapApplication from "shared/bootstrap";
import { getIsBootstrapped, getSignUpAllowed } from "shared/bootstrap/selectors";
import LoginView from "views/Login";
import SignUpView from "views/SignUp";
import { Backdrop, CircularProgress } from "@material-ui/core";

export class Root extends PureComponent {
  state = {
    loading: true,
  };

  static propTypes = {
    isReady: PropTypes.bool.isRequired,
    isAuthenticated: PropTypes.bool.isRequired,
    allowSignUp: PropTypes.bool.isRequired,
    bootstrapApplication: PropTypes.func.isRequired,
    bootstrapLogin: PropTypes.func.isRequired,
  };

  componentDidMount() {
    this.attemptBootstrap();
  }

  attemptBootstrap = () => {
    this.props.bootstrapApplication()
      .then(() => this.props.bootstrapLogin())
      .catch(error => {
        alert(error);
      })
      .finally(() => {
        this.setState({
          loading: false
        });
      });
  };

  renderUnauthenticated = () => {
    return (
      <Router>
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
      </Router>
    )
  };

  render() {
    const { isReady, isAuthenticated } = this.props;
    if (!isReady || this.state.loading) {
      return (
        <Backdrop  open={ true }>
          <CircularProgress color="inherit" />
        </Backdrop>
      );
    }

    if (!isAuthenticated) {
      return this.renderUnauthenticated();
    }

    return (
      <Router>
        <AuthenticatedApplication/>
      </Router>
    )
  }
}

export default connect(
  state => ({
    isAuthenticated: getIsAuthenticated(state),
    isReady: getIsBootstrapped(state),
    allowSignUp: getSignUpAllowed(state),
  }),
  dispatch => bindActionCreators({
    bootstrapApplication,
    bootstrapLogin,
  }, dispatch),
)(Root);
